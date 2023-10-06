package client

type RecordLogic interface {
	StartProcessRecords(countGoroutines int)
	StopProcessRecords()
}
