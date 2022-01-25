package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	connHost = "go-srv"
	connPort = "5500"
	connType = "tcp"
	dbServer = "db"
	port     = 3306
	user     = "root"
	password = "Isib1111"
	database = "dailyrecipe"

	REC_LIST       = 0
	REC_ING        = 1
	SAVE_ING       = 2
	DEL_ING        = 3
	SHOP_LIST      = 4
	SAVE_RECIPE    = 5
	ING_LIST       = 6
	UNIT_LIST      = 7
	DAY_LIST       = 8
	UPDATE_DAY_REC = 9
	RECIPE_AND_ALL = 10
)

type Day struct {
	Id   int
	Name string
}

type Unit struct {
	Id     int
	Name   string
	Symbol string
}

type Ingredient struct {
	Id        int
	Name      string
	Quantity  float64
	Unit      Unit
	ImagePath string
}

type Recipe struct {
	Id          int
	Name        string
	Multiplier  int
	Ingredients []Ingredient
	Description string
	ImagePath   string
	Day         Day
}

var db *sql.DB

func main() {
	//start listening
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	defer l.Close()

	//connect to db
	db, err = sql.Open("mysql", user+":"+password+"@tcp("+dbServer+":3306)/"+database)
	if err != nil {
		log.Fatal("Open connection to sql server failed:", err.Error())
	}
	fmt.Printf("Connected to sql server.\n")

	defer db.Close()
	//accept clients
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client connected.")

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	buffer, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}

	log.Println(string(buffer[:len(buffer)-1]))

	//query => [id, flag, arg]
	query := strings.Split(buffer, "\t")
	if len(query) == 3 {
		id := query[0]
		flagStr := query[1]
		arg := query[2]

		flag, err := strconv.Atoi(flagStr)
		if err != nil {
			fmt.Println("Bad request flag format")
			conn.Close()
			return
		}
		conn.Write([]byte(id + "\t"))
		switch flag {
		case REC_ING:
			{
				rId, err := strconv.Atoi(strings.Trim(arg, "\t\n"))
				if err != nil {
					fmt.Println("Bad request id format")
					conn.Close()
					return
				} else {
					sendRecipeIngredients(conn, rId)
				}
			}
		case REC_LIST:
			sendRecipeList(conn)
		case SAVE_RECIPE:
			saveRecipe(conn, arg)
		case SAVE_ING:
			saveIngredient(conn, arg)
		case ING_LIST:
			sendIngredientList(conn)
		case UNIT_LIST:
			sendUnitList(conn)
		case SHOP_LIST:
			sendShoppingList(conn)
		case DAY_LIST:
			sendDayList(conn)
		case UPDATE_DAY_REC:
			updateDayRecipe(conn, arg)
		case RECIPE_AND_ALL:
			sendRecipeAndAll(conn)
		}
		log.Println("Client message:", string(buffer[:len(buffer)-1]))
	} else {
		log.Print("Error. Query bad format : ")
		log.Println(query)
	}
	handleConnection(conn)
}

func sendRecipeAndAll(conn net.Conn) {
	rows, err := db.Query("select * from recipe_has_ingredient order by recipe_id")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var data [4]int

		err := rows.Scan(&data[0], &data[1], &data[2], &data[3])
		if err != nil {
			log.Fatal(err)
		}
		log.Println(data[0], data[1], data[2], data[3])
		marshalled, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func sendDayList(conn net.Conn) {
	rows, err := db.Query("select id, name from day")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var day Day
		err := rows.Scan(&day.Id, &day.Name)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(day.Id, day.Name)
		marshalled, err := json.Marshal(day)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func sendRecipeIngredients(conn net.Conn, recipeId int) {
	rows, err := db.Query("call select_recipe_ingredients(?)", recipeId)
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var ingredient Ingredient
		err := rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit.Id)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(ingredient.Id, ingredient.Name, ingredient.Quantity, ingredient.Unit.Id)
		marshalled, err := json.Marshal(ingredient)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	if buffer.Len() > 4 {
		buffer.Truncate(buffer.Len() - 1) //remove the last comma
	}
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func sendRecipeList(conn net.Conn) {
	rows, err := db.Query("select id, name, multiplier, description, image_path, day_id from recipe")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.Id, &recipe.Name, &recipe.Multiplier, &recipe.Description, &recipe.ImagePath, &recipe.Day.Id)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(recipe.Id, recipe.Name, recipe.ImagePath)
		marshalled, err := json.Marshal(recipe)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func sendIngredientList(conn net.Conn) {
	rows, err := db.Query("SELECT id, name, IFNULL(image_path, '') FROM ingredient")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var ingredient Ingredient
		err := rows.Scan(&ingredient.Id, &ingredient.Name, &ingredient.ImagePath)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(ingredient.Id, ingredient.Name, ingredient.ImagePath)
		marshalled, err := json.Marshal(ingredient)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

//arg : a recipe with ingredients to remove
func deleteIngredientFromRecipe(conn net.Conn, arg string) {
	var recipe Recipe
	err := json.Unmarshal([]byte(arg), &recipe)
	if err != nil {
		fmt.Println(err)
	}

	recipeId := strconv.Itoa(recipe.Id)
	var ingredientIds string
	for i := 0; i < len(recipe.Ingredients); i++ {
		ingredientIds += strconv.Itoa(recipe.Ingredients[i].Id) + ","
	}
	result, err := db.Query("call delete_recipe_ingredients(?,?)", recipeId, ingredientIds[:len(ingredientIds)-1])

	if err != nil || result == nil {
		fmt.Println(err)
	}
}

func sendShoppingList(conn net.Conn) {
	rows, err := db.Query("call shoppingList();")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var recipe Recipe
		var ingredient Ingredient
		var unit Unit
		err := rows.Scan(&recipe.Id, &ingredient.Id, &ingredient.Quantity, &unit.Id)
		if err != nil {
			log.Fatal(err)
		}
		recipe.Ingredients = append(recipe.Ingredients, ingredient)
		ingredient.Unit = unit

		log.Println(recipe.Id, ingredient.Id, ingredient.Quantity, unit.Id)
		marshalled, err := json.Marshal(recipe)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func sendUnitList(conn net.Conn) {
	rows, err := db.Query("SELECT id, name, symbol FROM unit")
	if err != nil {
		log.Fatal(err)
	}

	buffer := bytes.NewBufferString("[")
	for rows.Next() {
		var unit Unit
		err := rows.Scan(&unit.Id, &unit.Name, &unit.Symbol)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(unit.Id, unit.Name, unit.Symbol)
		marshalled, err := json.Marshal(unit)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(string(marshalled))
		buffer.WriteRune(',')
	}
	buffer.Truncate(buffer.Len() - 1) //remove the last comma
	buffer.WriteString("]")

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(buffer.String() + "\n"))
}

func saveIngredient(conn net.Conn, arg string) {
	var ingredient Ingredient
	var result int = 0
	err := json.Unmarshal([]byte(arg), &ingredient)
	if err != nil {
		fmt.Println(err)
		result++
	}
	if ingredientExist(ingredient) {
		row, err := db.Query("REPLACE INTO recipe (id, name, image_path) VALUE (?,?,?) WHERE id = '"+strconv.Itoa(ingredient.Id)+"'",
			strconv.Itoa(ingredient.Id), ingredient.Name, ingredient.ImagePath)
		if err != nil || row == nil {
			fmt.Println(err)
			result++
		}
	} else {
		row, err := db.Query("INSERT INTO recipe (name, image_path) VALUE (?,?)",
			ingredient.Name, ingredient.ImagePath)
		if err != nil || row == nil {
			fmt.Println(err)
			result++
		}
	}
	conn.Write([]byte(strconv.Itoa(result) + "\n"))
}

func ingredientExist(ingredient Ingredient) bool {
	return ingredient.Id != 0
}

func recipeExist(recipe Recipe) bool {
	return recipe.Id != 0
}

func saveRecipe(conn net.Conn, arg string) {
	var recipe Recipe
	var result int = 0
	var err = json.Unmarshal([]byte(arg), &recipe)
	if err != nil {
		fmt.Println(err)
		result++
	}
	var row *sql.Rows
	//insertion/update recipe
	if recipeExist(recipe) {
		row, err = db.Query("DELETE FROM recipe WHERE id = " + strconv.Itoa(recipe.Id))
		if err != nil || row == nil {
			fmt.Println(err)
			result++
		}
		//insert with same id
		row, err = db.Query("INSERT INTO recipe (id, name, multiplier, description, image_path, day_id) VALUE (?,?,?,?,?,?)", strconv.Itoa(recipe.Id), recipe.Name,
			recipe.Multiplier, recipe.Description, recipe.ImagePath, recipe.Day.Id)
		if err != nil || row == nil {
			fmt.Println(err)
			result++
		}
	} else {
		var id int64
		id, err = insert(db, recipe)
		recipe.Id = int(id)
		if err != nil {
			fmt.Println(err)
		}
	}

	//insertion ingredients that don't exist
	var buffer bytes.Buffer
	for _, ingredient := range recipe.Ingredients {
		buffer.WriteString(strconv.Itoa(ingredient.Id) + ",")
		row, err = db.Query("REPLACE INTO recipe_has_ingredient (recipe_id, ingredient_id, unit_id, quantity) VALUE (?,?,?,?)", strconv.Itoa(recipe.Id),
			strconv.Itoa(ingredient.Id), strconv.Itoa(ingredient.Unit.Id), strconv.FormatFloat(ingredient.Quantity, 'f', 3, 64))
		if err != nil || row == nil {
			fmt.Println(err)
			result++
		}
	}
	conn.Write([]byte(strconv.Itoa(recipe.Id) + "\n"))
}

func insert(db *sql.DB, recipe Recipe) (int64, error) {

	stmt, err := db.Prepare("INSERT INTO recipe (name, multiplier, description, image_path) VALUE (?,?,?,?)")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(recipe.Name, recipe.Multiplier, recipe.Description, recipe.ImagePath)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func updateDayRecipe(conn net.Conn, arg string) {
	var recipe Recipe
	var err = json.Unmarshal([]byte(arg), &recipe)
	if err != nil {
		fmt.Println(err)
	}
	stmt, err := db.Query("UPDATE recipe SET day_id = ? WHERE (id = ?)", recipe.Day.Id, recipe.Id)
	if err != nil {
		fmt.Println(err)
	}
	stmt.Close()

	conn.Write([]byte(strconv.Itoa(recipe.Id) + "\n"))
}
