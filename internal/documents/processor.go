package documents

type Processor interface {
	Extract(filePath string) (string, error)
	Create(content string, ouputPath string) (string, error)
}
