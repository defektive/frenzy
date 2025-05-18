package transformer

func replacerProcess(in []byte) ([]byte, error) {
	return in, nil
}

var Replacer = NewTransformer("replacer", "search and replace things", 1, replacerProcess)
