package firefly

func unmarshalTextToBytes(
	unmarshaller interface {
		UnmarshalText([]byte) error
		Bytes() []byte
	},
	source string,
) ([]byte, error) {
	err := unmarshaller.UnmarshalText([]byte(source))
	return unmarshaller.Bytes(), err
}
