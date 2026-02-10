package exporter

const (
	DefaultCsvDelimeter = ';'
)

type Options struct {
	CsvDelimeter rune
}

type Option func(*Options)

func CsvDelimeter(delim rune) Option {
	return func(o *Options) {
		o.CsvDelimeter = delim
	}
}
