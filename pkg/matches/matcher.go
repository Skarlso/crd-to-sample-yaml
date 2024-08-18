package matches

type Matcher interface {
	Match(sourceTemplateLocation string, payload []byte) error
}
