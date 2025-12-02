package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {
	// Cargar variables de entorno
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Advertencia: no se encontró el archivo .env")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI no está definido en el archivo .env")
	}

	fmt.Println("Conectando a MongoDB en URI:", uri)

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal("Error al conectar con MongoDB:", err)
		return nil
	}

	return client
}

func OpenCollection(nombreColeccion string, client *mongo.Client) *mongo.Collection {
	// Cargar variables de entorno
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Advertencia: no se encontró el archivo .env")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatal("DATABASE_NAME no está definido en el archivo .env")
	}

	fmt.Println("Usando la base de datos:", databaseName)

	coleccion := client.Database(databaseName).Collection(nombreColeccion)
	if coleccion == nil {
		log.Println("No se pudo abrir la colección:", nombreColeccion)
		return nil
	}

	return coleccion
}
