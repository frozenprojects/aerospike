# aerospike

An Aerospike database client where each set is called a "table" and each table has a struct assigned to it.

The main motivation for creating this client is to automatically build REST APIs from your structs using [aerogo/api](https://github.com/aerogo/api).

Struct fields must have a `json` tag if you want to save them in the database.

The lib also allows controlling Aerospike Go API directly by accessing `db.Client` in case you need low-level access.

## API

### NewDatabase

```go
db := aerospike.NewDatabase(
	"127.0.0.1",
	3000,
	"YOUR_NAMESPACE",

	// Register structs by giving a nil pointer to each struct.
	[]interface{}{
		(*User)(nil),
		(*Post)(nil),
		(*Thread)(nil),
	},
)
```

This will register 3 structs: User, Post and Thread.
The associated table names are automatically determined by the struct names: "User", "Post" and "Thread".

### Get

```go
obj, err := db.Get("User", "123")
user := obj.(*User)
```

### Set

```go
db.Set("User", "123", user)
```

Returns `error` and overwrites the object in the DB completely.

### Delete

```go
db.Delete("User", "123")
```

Returns `error`.

### All

```go
stream, err := db.All("User")
users := stream.(chan *User)

for user := range users {
	user.DoSomething()
}
```

Returns `(interface{}, error)` where the first parameter is a channel of the table's data type.

### Exists

```go
db.Exists("User", "123")
```

Returns `(bool, error)`.

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

```go
user, err := db.GetMap("User", "123")
firstName := user["firstName"]
```

Returns `(map[string]interface{}, error)`. The map contains the data for the retrieved object.

### GetObject

```go
type MyUser struct {
	FirstName string "json:`firstName`"
}

user := &MyUser{}
db.GetObject("User", "123", user)
fmt.Println(user.FirstName)
```

GetObject retrieves data from the table and stores it in the provided object. Unlike `db.Get()` the data type doesn't need to be pre-registered as a table.

### DeleteTable

```go
db.DeleteTable("User")
```

Deletes all content from the given table.

**Note:** It seems there is a bug in the official Aerospike client as the deleted data will show up on a cold start and allocate memory again.

### Type

```go
userType := db.Type("User")
```

Returns the type of the table.

### Types

```go
types := db.Types()
userType := types["User"]
```

Returns a `map[string]reflect.Type` of table to type relationships. This is used in [aerogo/api](https://github.com/aerogo/api) to automatically create a REST API from all of your struct data.

### Namespace

```go
namespace := db.Namespace()
```

Returns the previously registered namespace.

[![By Eduard Urbach](http://forthebadge.com/images/badges/built-with-love.svg)](https://github.com/blitzprog)