package main

import (
	"log"
	"strings"
	"time"
)

type TransactionLog struct {
	EventTime time.Time
	Action    TransactionAction
	Blame     string
	Success   bool
	Payload   string
}

type TransactionAction struct {
	ActionType  string
	ActionScope string
}

// generate a log to put into the System DB
func (s *SystemDB) generateTransactionLog(scope string, action string, success bool, payload string, User PublicAccessUser) {
	s.Transactions = append(s.Transactions, TransactionLog{
		EventTime: time.Now(),
		Action: TransactionAction{
			ActionType:  action,
			ActionScope: scope,
		},
		Blame:   User.Username,
		Success: success,
		Payload: payload,
	})
}

func (s *SystemDB) findTransactionLogs(args ...string) {
	for _, argItem := range args {
		log.Println(strings.Replace(argItem, " ", "", -1))
	}

	//log.Printf("%v", s.Transactions)
	log.Printf("%v transactions found.", len(s.Transactions))
}
