package tui

type dbLoadedMsg struct {
	dbName string
	count  int
	rows   [][]string
}

type dbCreatedMsg struct {
	dbName string
	count  int
	rows   [][]string
}

type errMsg struct {
	error
}
