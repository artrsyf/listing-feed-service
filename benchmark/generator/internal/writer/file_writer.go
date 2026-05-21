package writer

type FileWriter struct {
	OutputDir string

	CSV    *CSVWriter
	NDJSON *NDJSONWriter
}

func NewFileWriter(dir string) *FileWriter {
	return &FileWriter{
		OutputDir: dir,
	}
}

func (f *FileWriter) InitPostgres() error {
	users, _ := NewCSVWriter(f.OutputDir, "users.csv")
	orders, _ := NewCSVWriter(f.OutputDir, "orders.csv")
	items, _ := NewCSVWriter(f.OutputDir, "order_items.csv")
	products, _ := NewCSVWriter(f.OutputDir, "products.csv")
	categories, _ := NewCSVWriter(f.OutputDir, "categories.csv")

	f.CSV = users // placeholder pattern (will refactor per table later)

	_ = orders
	_ = items
	_ = products
	_ = categories

	return nil
}

func (f *FileWriter) InitElastic() error {
	nd, err := NewNDJSONWriter(f.OutputDir, "orders.ndjson")
	if err != nil {
		return err
	}

	f.NDJSON = nd
	return nil
}

func (f *FileWriter) Close() {
	if f.CSV != nil {
		f.CSV.Close()
	}
	if f.NDJSON != nil {
		f.NDJSON.Close()
	}
}
