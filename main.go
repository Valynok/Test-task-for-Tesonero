package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"strings"
)

type Users struct {
	Id int
	Name string `sql:"name"`
	Login string `sql:"login"`
	Password string `sql:"password"`
}

var PGdb *pg.DB



func main() {

	router := mux.NewRouter()

	pgOpt, err := pg.ParseURL("postgres://postgres:qweqwe@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	PGdb = pg.Connect(pgOpt)
	defer PGdb.Close()


	router.HandleFunc("/edit/{id}", EditHandler).Methods("PUT")
	router.HandleFunc("/", IndexHandler).Methods("GET")
	router.HandleFunc("/{id}", UserHandler).Methods("GET")
	router.HandleFunc("/create", CreateHandler).Methods("POST")
	router.HandleFunc("/delete/{id}", DeleteHandler).Methods("DELETE")

	http.ListenAndServe("localhost: 8080", router)

}

func IndexHandler (w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fmt.Fprintf(w, "All users: \n")

	var users []Users
	err := PGdb.Model(&users).Select()
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%+v", users)

}

func UserHandler (w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestId, err := strconv.Atoi(mux.Vars(r)["id"])

	user := &Users{Id: requestId}
	err = PGdb.Select(user)
	if (err != nil) && (err.Error() == "pg: no rows in result set") {
		http.Error(w, "User with this ID didn't found", 400)
		return
	} else if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%+v", user)
}



func CreateHandler (w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()


	decoder := json.NewDecoder(r.Body)
	var u Users
	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}

	Validate(u, w)

	err = PGdb.Insert(&u)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "User created.")
}

func DeleteHandler (w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestId, err := strconv.Atoi(mux.Vars(r)["id"])

	user := &Users{Id: requestId}
	err = PGdb.Delete(user)
	NoUser(err, w)
	fmt.Fprintf(w, "Deleted user with ID %v\n", requestId)
}


func EditHandler (w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestId, err := strconv.Atoi(mux.Vars(r)["id"])

	user := &Users{Id: requestId}
	err = PGdb.Select(user)
	NoUser(err, w)


	decoder := json.NewDecoder(r.Body)
	var u Users
	err = decoder.Decode(&u)
	if err != nil {
		panic(err)
	}

	Validate(u, w)

	user.Name = u.Name
	user.Password = u.Password
	user.Login = u.Login
	err = PGdb.Update(user)
	if err != nil {
		panic(err)
	}


}

func NoUser (err error, w http.ResponseWriter) {
	if (err != nil) && (err.Error() == "pg: no rows in result set") {
		http.Error(w, "User with this ID didn't found", 400)
		return
	} else if err != nil {
		panic(err)
	}
}

func Validate(u Users, w http.ResponseWriter) {
	fields := make([]string, 0)

	if u.Name == "" {
		fields = append(fields, "name")
	}
	if u.Login == "" {
		fields = append(fields, "login")
	}
	if u.Password == "" {
		fields = append(fields, "password")
	}

	if (len(fields) != 0) {
		str := "You didn't fill next fields: "+strings.Join(fields, ", ")
		http.Error(w, str, 400)
	}
}