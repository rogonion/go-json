package internal

type PathSegmentsIndexes struct {
	CurrentRecursive  int
	LastRecursive     int
	CurrentCollection int
	LastCollection    int
}

type TestData struct {
	TestTitle                string
	LogErrorsIfExpectedNotOk bool
}
