# aerospike
Aerospike database client.

## API
### NewDatabase
```go
db := aerospike.NewDatabase(
	"127.0.0.1",
	3000,
	"YOUR_NAMESPACE",
	[]interface{}{
		// Class names = Table names
		new(User),
		new(Post),
		new(Thread),
	},
)
```
### Get
```go
obj, err := db.Get("User", "123")
user := obj.(*User)
```
### Set
```go
db.Set("User", "123", user)
```
### Delete
```go
db.Delete("User", "123")
```
### All
```go
db.All("User")
```
### GetMany
```go
objects, err := db.GetMany("User", []string{
	"123",
	"456",
	"789",
})

users := objects.([]*User)
```
### GetMap
### GetObject
### DeleteTable
### Namespace