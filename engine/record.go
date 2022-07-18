package engine

/**
 * Record : An interface - a referenceable type - for the Bleve "single-table style" store's record
 */

type RecordType string

const (
	RecordTypeField string     = "_type"
	RecordSource    RecordType = "source"
	RecordResource  RecordType = "resource"
)

type Record map[string]interface{}

func (record Record) SetAll(entries map[string]interface{}) Record {
	for key, value := range entries {
		record[key] = value
	}

	return record
}

func (record Record) SetType(recordType RecordType) Record {
	return record.SetAll(map[string]interface{}{
		RecordTypeField: recordType,
	})
}
