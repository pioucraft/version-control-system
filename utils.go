package main

type Change struct {
	Op string
	Value string
	Line int
}

type SimpleCommitStruct struct {
	Key string
	OldContent string
	NewContent string
	BinaryContent []byte
}

func IsBinary(data []byte) bool {
	maxCheck := 8000
	if len(data) < maxCheck {
		maxCheck = len(data)
	}
	for i := 0; i < maxCheck; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
