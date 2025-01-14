/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package auditdb

import (
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	view2 "github.com/hyperledger-labs/fabric-smart-client/platform/view"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/services/flogging"
	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/auditor/auditdb/driver"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

var logger = flogging.MustGetLogger("token-sdk.auditor.auditdb")

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.Driver)
)

// Register makes a AuditDB driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver driver.Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("auditor: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("auditor: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func unregisterAllDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()
	// For tests.
	drivers = make(map[string]driver.Driver)
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// TxStatus is the status of a transaction
type TxStatus string

const (
	// Pending is the status of a transaction that has been submitted to the ledger
	Pending TxStatus = "Pending"
	// Confirmed is the status of a transaction that has been confirmed by the ledger
	Confirmed TxStatus = "Confirmed"
	// Deleted is the status of a transaction that has been deleted due to a failure to commit
	Deleted TxStatus = "Deleted"
)

// TransactionType is the type of transaction
type TransactionType int

const (
	// Issue is the type of transaction for issuing assets
	Issue TransactionType = iota
	// Transfer is the type of transaction for transferring assets
	Transfer
	// Redeem is the type of transaction for redeeming assets
	Redeem
)

// MovementRecord is a record of a movement of assets
type MovementRecord struct {
	// TxID is the transaction ID
	TxID string
	// EnrollmentID is the enrollment ID of the account that is receiving or sendeing
	EnrollmentID string
	// TokenType is the type of token
	TokenType string
	// Amount is positive if tokens are received. Negative otherwise
	Amount *big.Int
	// Status is the status of the transaction
	Status TxStatus
}

// TransactionRecord is a record of a transaction
type TransactionRecord struct {
	// TxID is the transaction ID
	TxID string
	// TransactionType is the type of transaction
	TransactionType TransactionType
	// SenderEID is the enrollment ID of the account that is sending tokens
	SenderEID string
	// RecipientEID is the enrollment ID of the account that is receiving tokens
	RecipientEID string
	// TokenType is the type of token
	TokenType string
	// Amount is positive if tokens are received. Negative otherwise
	Amount *big.Int
	// Timestamp is the time the transaction was submitted to the auditor
	Timestamp time.Time
	// Status is the status of the transaction
	Status TxStatus
}

func (t *TransactionRecord) String() string {
	var s strings.Builder
	s.WriteString("{")
	s.WriteString(t.TxID)
	s.WriteString(" ")
	s.WriteString(strconv.Itoa(int(t.TransactionType)))
	s.WriteString(" ")
	s.WriteString(t.SenderEID)
	s.WriteString(" ")
	s.WriteString(t.RecipientEID)
	s.WriteString(" ")
	s.WriteString(t.TokenType)
	s.WriteString(" ")
	s.WriteString(t.Amount.String())
	s.WriteString(" ")
	s.WriteString(t.Timestamp.String())
	s.WriteString(" ")
	s.WriteString(string(t.Status))
	s.WriteString("}")
	return s.String()
}

// TransactionIterator is an iterator over transaction records
type TransactionIterator struct {
	it driver.TransactionIterator
}

// Close closes the iterator. It must be called when done with the iterator.
func (t *TransactionIterator) Close() {
	t.it.Close()
}

// Next returns the next transaction record, if any.
// It returns nil, nil if there are no more records.
func (t *TransactionIterator) Next() (*TransactionRecord, error) {
	next, err := t.it.Next()
	if err != nil {
		return nil, err
	}
	if next == nil {
		return nil, nil
	}
	return &TransactionRecord{
		TxID:            next.TxID,
		TransactionType: TransactionType(next.TransactionType),
		SenderEID:       next.SenderEID,
		RecipientEID:    next.RecipientEID,
		TokenType:       next.TokenType,
		Amount:          next.Amount,
		Timestamp:       next.Timestamp,
		Status:          TxStatus(next.Status),
	}, nil
}

// QueryExecutor executors queries against the audit DB
type QueryExecutor struct {
	db     *AuditDB
	closed bool
}

// NewPaymentsFilter returns a programmable filter over the payments sent or received by enrollment IDs.
func (qe *QueryExecutor) NewPaymentsFilter() *PaymentsFilter {
	return &PaymentsFilter{
		db: qe.db,
	}
}

// NewHoldingsFilter returns a programmable filter over the holdings owned by enrollment IDs.
func (qe *QueryExecutor) NewHoldingsFilter() *HoldingsFilter {
	return &HoldingsFilter{
		db: qe.db,
	}
}

// Transactions returns an iterators of transaction records in the given time internal.
// If from and to are both nil, all transactions are returned.
func (qe *QueryExecutor) Transactions(from, to *time.Time) (*TransactionIterator, error) {
	it, err := qe.db.db.QueryTransactions(from, to)
	if err != nil {
		return nil, errors.Errorf("failed to query transactions: %s", err)
	}
	return &TransactionIterator{it: it}, nil
}

// Done closes the query executor. It must be called when the query executor is no longer needed.s
func (qe *QueryExecutor) Done() {
	if qe.closed {
		return
	}
	qe.db.counter.Dec()
	qe.db.storeLock.RUnlock()
	qe.closed = true
}

// AuditDB is a database that stores audit information
type AuditDB struct {
	counter atomic.Int32

	// the vault handles access concurrency to the store using storeLock.
	// In particular:
	// * when a directQueryExecutor is returned, it holds a read-lock;
	//   when Done is called on it, the lock is released.
	// * when an interceptor is returned (using NewRWSet (in case the
	//   transaction context is generated from nothing) or GetRWSet
	//   (in case the transaction context is received from another node)),
	//   it holds a read-lock; when Done is called on it, the lock is released.
	// * an exclusive lock is held when Commit is called.
	db        driver.AuditDB
	storeLock sync.RWMutex

	eIDsLocks sync.Map

	// status related fields
	statusUpdating atomic.Bool
	pendingTXs     []string
	wg             sync.WaitGroup
}

func newAuditDB(p driver.AuditDB) *AuditDB {
	return &AuditDB{
		db:         p,
		eIDsLocks:  sync.Map{},
		pendingTXs: make([]string, 0, 10000),
	}
}

// Append appends the passed token request to the audit database
func (db *AuditDB) Append(req *token.Request) error {
	logger.Debugf("Appending new record... [%d]", db.counter)
	db.storeLock.Lock()
	defer db.storeLock.Unlock()
	logger.Debug("lock acquired")

	record, err := req.AuditRecord()
	if err != nil {
		return errors.WithMessagef(err, "failed getting audit records for request [%s]", req.Anchor)
	}

	if err := db.db.BeginUpdate(); err != nil {
		db.rollback(err)
		return errors.WithMessagef(err, "begin update for txid '%s' failed", record.Anchor)
	}
	if err := db.appendSendMovements(record); err != nil {
		db.rollback(err)
		return errors.WithMessagef(err, "append send movements for txid '%s' failed", record.Anchor)
	}
	if err := db.appendReceivedMovements(record); err != nil {
		db.rollback(err)
		return errors.WithMessagef(err, "append received movements for txid '%s' failed", record.Anchor)
	}
	if err := db.appendTransactions(record); err != nil {
		db.rollback(err)
		return errors.WithMessagef(err, "append transactions for txid '%s' failed", record.Anchor)
	}
	if err := db.db.Commit(); err != nil {
		db.rollback(err)
		return errors.WithMessagef(err, "committing tx for txid '%s' failed", record.Anchor)
	}

	logger.Debugf("Appending new completed without errors")
	return nil
}

// NewQueryExecutor returns a new query executor
func (db *AuditDB) NewQueryExecutor() *QueryExecutor {
	db.counter.Inc()
	db.storeLock.RLock()

	return &QueryExecutor{db: db}
}

// SetStatus sets the status of the audit records with the passed transaction id to the passed status
func (db *AuditDB) SetStatus(txID string, status TxStatus) error {
	logger.Debugf("Set status [%s][%s]...[%d]", txID, status, db.counter)
	db.storeLock.Lock()
	defer db.storeLock.Unlock()
	logger.Debug("lock acquired")

	if err := db.db.SetStatus(txID, driver.TxStatus(status)); err != nil {
		db.rollback(err)
		return errors.Wrapf(err, "failed setting status [%s][%s]", txID, status)
	}
	logger.Debugf("Set status [%s][%s]...[%d] done without errors", txID, status, db.counter)
	return nil
}

// AcquireLocks acquires locks for the passed enrollment ids.
// This can be used to prevent concurrent read/write access to the audit records of the passed enrollment ids.
func (db *AuditDB) AcquireLocks(eIDs ...string) error {
	for _, id := range deduplicate(eIDs) {
		lock, _ := db.eIDsLocks.LoadOrStore(id, &sync.RWMutex{})
		lock.(*sync.RWMutex).Lock()
	}
	return nil
}

// Unlock unlocks the locks for the passed enrollment ids.
func (db *AuditDB) Unlock(eIDs ...string) {
	for _, id := range deduplicate(eIDs) {
		lock, ok := db.eIDsLocks.Load(id)
		if !ok {
			logger.Warnf("unlock for enrollment id [%s] not possible, lock never acquired", id)
			continue
		}
		lock.(*sync.RWMutex).Unlock()
	}
}

func (db *AuditDB) appendSendMovements(record *token.AuditRecord) error {
	inputs := record.Inputs
	outputs := record.Outputs
	// we need to consider both inputs and outputs enrollment IDs because the record can refer to a redeem
	eIDs := joinIOEIDs(record)
	tokenTypes := outputs.TokenTypes()

	for _, eID := range eIDs {
		for _, tokenType := range tokenTypes {
			sent := inputs.ByEnrollmentID(eID).ByType(tokenType).Sum().ToBigInt()
			received := outputs.ByEnrollmentID(eID).ByType(tokenType).Sum().ToBigInt()
			diff := sent.Sub(sent, received)
			if diff.Cmp(big.NewInt(0)) <= 0 {
				continue
			}

			if err := db.db.AddMovement(&driver.MovementRecord{
				TxID:         record.Anchor,
				EnrollmentID: eID,
				Amount:       diff.Neg(diff),
				TokenType:    tokenType,
				Status:       driver.Pending,
			}); err != nil {
				if err1 := db.db.Discard(); err1 != nil {
					logger.Errorf("got error %s; discarding caused %s", err.Error(), err1.Error())
				}
				return err
			}
		}
	}
	logger.Debugf("finished to append send movements for tx [%s]", record.Anchor)

	return nil
}

func (db *AuditDB) appendReceivedMovements(record *token.AuditRecord) error {
	inputs := record.Inputs
	outputs := record.Outputs
	// we need to consider both inputs and outputs enrollment IDs because the record can refer to a redeem
	eIDs := joinIOEIDs(record)
	tokenTypes := outputs.TokenTypes()

	for _, eID := range eIDs {
		for _, tokenType := range tokenTypes {
			received := outputs.ByEnrollmentID(eID).ByType(tokenType).Sum().ToBigInt()
			sent := inputs.ByEnrollmentID(eID).ByType(tokenType).Sum().ToBigInt()
			diff := received.Sub(received, sent)
			if diff.Cmp(big.NewInt(0)) <= 0 {
				// Nothing received
				continue
			}

			if err := db.db.AddMovement(&driver.MovementRecord{
				TxID:         record.Anchor,
				EnrollmentID: eID,
				Amount:       diff,
				TokenType:    tokenType,
				Status:       driver.Pending,
			}); err != nil {
				if err1 := db.db.Discard(); err1 != nil {
					logger.Errorf("got error %s; discarding caused %s", err.Error(), err1.Error())
				}
				return err
			}
		}
	}
	logger.Debugf("finished to append received movements for tx [%s]", record.Anchor)

	return nil
}

func (db *AuditDB) appendTransactions(record *token.AuditRecord) error {
	inputs := record.Inputs
	outputs := record.Outputs

	actionIndex := 0
	timestamp := time.Now()
	for {
		// collect inputs and outputs from the same action
		ins := inputs.Filter(func(t *token.Input) bool {
			return t.ActionIndex == actionIndex
		})
		ous := outputs.Filter(func(t *token.Output) bool {
			return t.ActionIndex == actionIndex
		})
		if ins.Count() == 0 && ous.Count() == 0 {
			logger.Debugf("no actions left for tx [%s][%d]", record.Anchor, actionIndex)
			// no more actions
			break
		}

		// create a transaction record from ins and ous

		// All ins should be for same EID, check this
		inEIDs := ins.EnrollmentIDs()
		if len(inEIDs) > 1 {
			return errors.Errorf("expected at most 1 input enrollment id, got %d", len(inEIDs))
		}
		inEID := ""
		if len(inEIDs) == 1 {
			inEID = inEIDs[0]
		}

		outEIDs := ous.EnrollmentIDs()
		outEIDs = append(outEIDs, "")
		outTT := ous.TokenTypes()
		for _, outEID := range outEIDs {
			for _, tokenType := range outTT {
				received := outputs.ByEnrollmentID(outEID).ByType(tokenType).Sum().ToBigInt()
				if received.Cmp(big.NewInt(0)) <= 0 {
					continue
				}

				tt := driver.Issue
				if len(inEIDs) != 0 {
					if len(outEID) == 0 {
						tt = driver.Redeem
					} else {
						tt = driver.Transfer
					}
				}

				if err := db.db.AddTransaction(&driver.TransactionRecord{
					TxID:            record.Anchor,
					SenderEID:       inEID,
					RecipientEID:    outEID,
					TokenType:       tokenType,
					Amount:          received,
					Status:          driver.Pending,
					TransactionType: tt,
					Timestamp:       timestamp,
				}); err != nil {
					if err1 := db.db.Discard(); err1 != nil {
						logger.Errorf("got error %s; discarding caused %s", err.Error(), err1.Error())
					}
					return err
				}
			}
		}

		actionIndex++
	}
	logger.Debugf("finished appending transactions for tx [%s]", record.Anchor)

	return nil
}

func (db *AuditDB) rollback(err error) {
	if err1 := db.db.Discard(); err1 != nil {
		logger.Errorf("got error %s; discarding caused %s", err.Error(), err1.Error())
	}
}

// Manager handles the audit databases
type Manager struct {
	sp     view2.ServiceProvider
	driver string
	mutex  sync.Mutex
	dbs    map[string]*AuditDB
}

// NewManager creates a new audit manager
func NewManager(sp view2.ServiceProvider, driver string) *Manager {
	return &Manager{
		sp:     sp,
		driver: driver,
		dbs:    map[string]*AuditDB{},
	}
}

// AuditDB returns an AuditDB for the given auditor wallet
func (cm *Manager) AuditDB(w *token.AuditorWallet) (*AuditDB, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	id := w.ID()
	c, ok := cm.dbs[id]
	if !ok {
		driver, err := drivers[cm.driver].Open(cm.sp, "")
		if err != nil {
			return nil, errors.Wrapf(err, "failed instantiating audit db driver")
		}
		c = newAuditDB(driver)
		cm.dbs[id] = c
	}
	return c, nil
}

var (
	managerType = reflect.TypeOf((*Manager)(nil))
)

// GetAuditDB returns the AuditDB for the given auditor wallet.
// Nil might be returned if the auditor wallet is not found or an error occurred.
func GetAuditDB(sp view2.ServiceProvider, w *token.AuditorWallet) *AuditDB {
	if w == nil {
		logger.Debugf("no auditor wallet provided")
		return nil
	}
	s, err := sp.GetService(managerType)
	if err != nil {
		logger.Errorf("failed to get audit manager service: [%s]", err)
		return nil
	}
	c, err := s.(*Manager).AuditDB(w)
	if err != nil {
		logger.Errorf("failed to get audit db for wallet [%s]: [%s]", w.ID(), err)
		return nil
	}
	return c
}

// joinIOEIDs joins enrollment IDs of inputs and outputs
func joinIOEIDs(record *token.AuditRecord) []string {
	iEIDs := record.Inputs.EnrollmentIDs()
	oEIDs := record.Outputs.EnrollmentIDs()
	eIDs := append(iEIDs, oEIDs...)
	eIDs = deduplicate(eIDs)
	return eIDs
}

// deduplicate removes duplicate entries from a slice
func deduplicate(source []string) []string {
	support := make(map[string]bool)
	var res []string
	for _, item := range source {
		if _, value := support[item]; !value {
			support[item] = true
			res = append(res, item)
		}
	}
	return res
}
