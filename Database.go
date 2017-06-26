package aerospike

import (
	"errors"
	"reflect"

	as "github.com/aerospike/aerospike-client-go"
)

func init() {
	// This will make Aerospike use json tags for the field names in the database.
	as.SetAerospikeTag("json")
}

// Database represents the Aerospike database.
type Database struct {
	namespace string
	types     map[string]reflect.Type
	Client    *as.Client
}

// NewDatabase creates a new database client.
func NewDatabase(host string, port int, namespace string, tables []interface{}) *Database {
	// Convert example objects to their respective types
	tableTypes := make(map[string]reflect.Type)
	for _, example := range tables {
		typeInfo := reflect.TypeOf(example).Elem()
		tableTypes[typeInfo.Name()] = typeInfo
	}

	// Client policy
	clientPolicy := as.NewClientPolicy()
	clientPolicy.ConnectionQueueSize = 1024

	// Create client
	client, err := as.NewClientWithPolicy(clientPolicy, host, port)

	if err != nil {
		panic(err)
	}

	// Make Set() calls delete old fields instead of only updating new ones
	client.DefaultWritePolicy.RecordExistsAction = as.REPLACE

	// This will make delete actually...delete things...you know.
	// Otherwise they'll just reappear after a node restart.
	// client.DefaultWritePolicy.DurableDelete = true

	// Make scans faster
	client.DefaultScanPolicy.Priority = as.HIGH
	client.DefaultScanPolicy.ConcurrentNodes = true
	client.DefaultScanPolicy.IncludeBinData = true

	return &Database{
		namespace: namespace,
		types:     tableTypes,
		Client:    client,
	}
}

// Get retrieves an object from the table.
func (db *Database) Get(table string, id string) (interface{}, error) {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return nil, keyErr
	}

	t, exists := db.types[table]

	if !exists {
		return nil, errors.New("Data type has not been defined for table " + table)
	}

	obj := reflect.New(t).Interface()
	err := db.Client.GetObject(nil, pk, obj)

	return obj, err
}

// Set sets an object's data for the given ID and erases old fields.
func (db *Database) Set(table string, id string, obj interface{}) error {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return keyErr
	}

	return db.Client.PutObject(nil, pk, obj)
}

// Delete deletes an object from the database and returns if it existed.
func (db *Database) Delete(table string, id string) (existed bool, err error) {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return false, keyErr
	}

	return db.Client.Delete(nil, pk)
}

// Exists tells you if the given record exists.
func (db *Database) Exists(table string, id string) (bool, error) {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return false, keyErr
	}

	return db.Client.Exists(nil, pk)
}

// Scan writes all objects from a given table to the channel.
func (db *Database) Scan(table string, channel interface{}) error {
	_, err := db.Client.ScanAllObjects(nil, channel, db.namespace, table)
	return err
}

// All returns a stream of all objects in the given table.
func (db *Database) All(table string) (interface{}, error) {
	channel := reflect.MakeChan(db.types[table], 0)
	err := db.Scan(table, channel)
	return channel, err
}

// GetObject retrieves data from the table and stores it in the provided object.
func (db *Database) GetObject(table string, id string, obj interface{}) error {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return keyErr
	}

	return db.Client.GetObject(nil, pk, obj)
}

// GetMap retrieves the data as a map[string]interface{}.
func (db *Database) GetMap(table string, id string) (as.BinMap, error) {
	pk, keyErr := as.NewKey(db.namespace, table, id)

	if keyErr != nil {
		return nil, keyErr
	}

	rec, err := db.Client.Get(nil, pk)

	if err != nil {
		return nil, err
	}

	if rec == nil {
		return nil, errors.New("Record not found")
	}

	return rec.Bins, nil
}

// GetMany performs a Get request for every ID in the ID list and returns a slice of objects.
func (db *Database) GetMany(table string, idList []string) (interface{}, error) {
	// Get data type for that table
	t, exists := db.types[table]

	if !exists {
		return nil, errors.New("Data type has not been defined for table " + table)
	}

	// Number of keys
	num := len(idList)

	// Create a slice of pointers
	objType := reflect.SliceOf(t)
	ptrType := reflect.SliceOf(reflect.PtrTo(t))
	objects := reflect.MakeSlice(objType, num, num)
	pointers := reflect.MakeSlice(ptrType, num, num)

	// Return early if there's nothing to do
	if num == 0 {
		return pointers.Interface(), nil
	}

	keys := make([]*as.Key, num, num)
	interfaceSlice := make([]interface{}, num, num)

	for i := 0; i < num; i++ {
		keys[i], _ = as.NewKey(db.namespace, table, idList[i])

		objAddr := objects.Index(i).Addr()
		pointers.Index(i).Set(objAddr)
		interfaceSlice[i] = objAddr.Interface()
	}

	// This needs an interface slice of pointers to structs.
	_, err := db.Client.BatchGetObjects(nil, keys, interfaceSlice)

	if err != nil {
		return nil, err
	}

	return pointers.Interface(), nil
}

// DeleteTable deletes all content from the given table.
func (db *Database) DeleteTable(table string) error {
	return db.Client.Truncate(nil, db.namespace, table, nil)
}

// Namespace returns the name of the namespace.
func (db *Database) Namespace() string {
	return db.namespace
}

// Type returns the type of the table.
func (db *Database) Type(table string) reflect.Type {
	return db.types[table]
}

// Types returns the types of all tables as a map.
func (db *Database) Types() map[string]reflect.Type {
	return db.types
}

// // ForEach ...
// func ForEach(set string, callback func(as.BinMap)) {
// 	recs, _ := client.ScanAll(scanPolicy, namespace, set)

// 	for res := range recs.Results() {
// 		if res.Err != nil {
// 			recs.Close()
// 			return
// 		}

// 		callback(res.Record.Bins)
// 	}

// 	recs.Close()
// }
