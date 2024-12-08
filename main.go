package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Variables globales
var clienteMongo *mongo.Client
var baseDatos *mongo.Database

// Conexión a MongoDB
func conectarMongo() {
	uri := "mongodb+srv://alejcaa1109:7jxDqqLEOMw94VwK@cluster0.guer8.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	var err error
	clienteMongo, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error al crear cliente MongoDB:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = clienteMongo.Connect(ctx)
	if err != nil {
		log.Fatal("Error al conectar a MongoDB:", err)
	}
	baseDatos = clienteMongo.Database("parqueadero_autos_colombia")
	fmt.Println("¡Conectado exitosamente a MongoDB!")
}

// Estructuras de datos
type Pago struct {
	ID          primitive.ObjectID
	UsuarioID   string
	VehiculoID  string
	Monto       float64
	Fecha       time.Time
	Descripcion string
}

// Rutas principales
func manejarRutas() {
	http.HandleFunc("/pagos", gestionarPagos)
	fmt.Println("Servidor iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Gestión de pagos
func gestionarPagos(w http.ResponseWriter, r *http.Request) {
	coleccion := baseDatos.Collection("pagos")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch r.Method {
	case "GET":
		var pagos []Pago
		cursor, err := coleccion.Find(ctx, bson.M{})
		if err != nil {
			http.Error(w, "Error al obtener pagos", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var pago Pago
			if err := cursor.Decode(&pago); err != nil {
				http.Error(w, "Error al decodificar pago", http.StatusInternalServerError)
				return
			}
			pagos = append(pagos, pago)
		}
		json.NewEncoder(w).Encode(pagos)

	case "POST":
		var nuevoPago Pago
		if err := json.NewDecoder(r.Body).Decode(&nuevoPago); err != nil {
			http.Error(w, "Datos inválidos", http.StatusBadRequest)
			return
		}
		nuevoPago.Fecha = time.Now()
		result, err := coleccion.InsertOne(ctx, nuevoPago)
		if err != nil {
			http.Error(w, "Error al crear el pago", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)

	case "PUT":
		var pagoActualizado Pago
		if err := json.NewDecoder(r.Body).Decode(&pagoActualizado); err != nil {
			http.Error(w, "Datos inválidos", http.StatusBadRequest)
			return
		}
		id, _ := primitive.ObjectIDFromHex(pagoActualizado.ID.Hex())
		_, err := coleccion.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": pagoActualizado})
		if err != nil {
			http.Error(w, "Error al actualizar el pago", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode("Pago actualizado correctamente")

	case "DELETE":
		id := r.URL.Query().Get("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		_, err = coleccion.DeleteOne(ctx, bson.M{"_id": objectID})
		if err != nil {
			http.Error(w, "Error al eliminar el pago", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode("Pago eliminado correctamente")

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// Función principal
func main() {
	conectarMongo()
	manejarRutas()
}
