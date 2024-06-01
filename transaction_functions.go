package main

type TransactionLog struct {
	EventTime string
	Action    TransactionAction
	Blame     string
}

type TransactionAction struct {
	ActionType  string
	ActionScope string
}
