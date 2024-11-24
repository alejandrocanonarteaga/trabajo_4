package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Conexión a MongoDB
func connectToMongo() {
	uri := "mongodb+srv://alejcaa1109:7jxDqqLEOMw94VwK@cluster0.guer8.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("¡Conectado exitosamente a MongoDB!")
}

// Modelos para la base de datos
type Usuario struct {
	ID       string `json:"id,omitempty"`
	Nombre   string `json:"nombre,omitempty"`
	Correo   string `json:"correo,omitempty"`
	Telefono string `json:"telefono,omitempty"`
}

type Vehiculo struct {
	Placa        string    `json:"placa,omitempty"`
	UsuarioID    string    `json:"usuario_id,omitempty"`
	FechaEntrada time.Time `json:"fecha_entrada,omitempty"`
	FechaSalida  time.Time `json:"fecha_salida,omitempty"`
}

type Celda struct {
	ID          string `json:"id,omitempty"`
	Descripcion string `json:"descripcion,omitempty"`
	Estado      string `json:"estado,omitempty"` // Disponible / Ocupado
	VehiculoID  string `json:"vehiculo_id,omitempty"`
}

type Pago struct {
	ID         string    `json:"id,omitempty"`
	VehiculoID string    `json:"vehiculo_id,omitempty"`
	Monto      float64   `json:"monto,omitempty"`
	FechaPago  time.Time `json:"fecha_pago,omitempty"`
}

// Gestión de Usuarios
func createUsuario(w http.ResponseWriter, r *http.Request) {
	var usuario Usuario
	_ = json.NewDecoder(r.Body).Decode(&usuario)

	collection := client.Database("autos_colombia").Collection("usuarios")
	_, err := collection.InsertOne(context.Background(), usuario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(usuario)
}

func getUsuarios(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("autos_colombia").Collection("usuarios")

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var usuarios []Usuario
	for cursor.Next(context.Background()) {
		var usuario Usuario
		if err := cursor.Decode(&usuario); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		usuarios = append(usuarios, usuario)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarios)
}

func updateUsuario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var usuario Usuario
	_ = json.NewDecoder(r.Body).Decode(&usuario)

	collection := client.Database("autos_colombia").Collection("usuarios")
	filter := bson.M{"id": id}
	update := bson.M{"$set": usuario}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Usuario actualizado exitosamente"})
}

// Gestión de Vehículos
func createVehiculo(w http.ResponseWriter, r *http.Request) {
	var vehiculo Vehiculo
	_ = json.NewDecoder(r.Body).Decode(&vehiculo)
	vehiculo.FechaEntrada = time.Now()

	collection := client.Database("autos_colombia").Collection("vehiculos")
	_, err := collection.InsertOne(context.Background(), vehiculo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vehiculo)
}

func updateVehiculoSalida(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	placa := params["placa"]

	collection := client.Database("autos_colombia").Collection("vehiculos")
	filter := bson.M{"placa": placa}
	update := bson.M{"$set": bson.M{"fecha_salida": time.Now()}}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Salida registrada exitosamente"})
}

// Gestión de Celdas
func createCelda(w http.ResponseWriter, r *http.Request) {
	var celda Celda
	_ = json.NewDecoder(r.Body).Decode(&celda)
	celda.Estado = "Disponible"

	collection := client.Database("autos_colombia").Collection("celdas")
	_, err := collection.InsertOne(context.Background(), celda)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(celda)
}

func updateCeldaEstado(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var update struct {
		Estado     string `json:"estado,omitempty"`
		VehiculoID string `json:"vehiculo_id,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&update)

	collection := client.Database("autos_colombia").Collection("celdas")
	filter := bson.M{"id": id}
	updateFields := bson.M{"$set": update}

	_, err := collection.UpdateOne(context.Background(), filter, updateFields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Celda actualizada exitosamente"})
}

// Gestión de Pagos
func createPago(w http.ResponseWriter, r *http.Request) {
	var pago Pago
	_ = json.NewDecoder(r.Body).Decode(&pago)
	pago.FechaPago = time.Now()

	collection := client.Database("autos_colombia").Collection("pagos")
	_, err := collection.InsertOne(context.Background(), pago)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pago)
}

// Funciones para obtener informacion de Vehiculos y Celdas.
func getVehiculos(w http.ResponseWriter, r *http.Request) {
	vehiculos := []Vehiculo{
		{Placa: "ABC123", UsuarioID: "123", FechaEntrada: time.Now(), FechaSalida: time.Now()},
		{Placa: "XYZ456", UsuarioID: "456", FechaEntrada: time.Now(), FechaSalida: time.Now()},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehiculos)
}

func getCeldas(w http.ResponseWriter, r *http.Request) {
	celdas := []Celda{
		{ID: "1", Estado: "Disponible", VehiculoID: ""},
		{ID: "2", Estado: "Ocupado", VehiculoID: "ABC123"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(celdas)
}

// Función principal
func main() {
	connectToMongo()
	defer client.Disconnect(context.Background())

	router := mux.NewRouter()

	// Rutas de usuarios
	router.HandleFunc("/usuarios", createUsuario).Methods("POST")
	router.HandleFunc("/usuarios", getUsuarios).Methods("GET")
	router.HandleFunc("/usuarios/{id}", updateUsuario).Methods("PUT")

	// Rutas de vehículos
	router.HandleFunc("/vehiculos", createVehiculo).Methods("POST")
	router.HandleFunc("/vehiculos/{placa}/salida", updateVehiculoSalida).Methods("PUT")

	// Rutas de celdas
	router.HandleFunc("/celdas", createCelda).Methods("POST")
	router.HandleFunc("/celdas/{id}/estado", updateCeldaEstado).Methods("PUT")

	// Rutas de pagos
	router.HandleFunc("/pagos", createPago).Methods("POST")

	// Rutas de consulta
	router.HandleFunc("/vehiculos", getVehiculos).Methods("GET")
	router.HandleFunc("/celdas", getCeldas).Methods("GET")

	// Servir la API
	log.Fatal(http.ListenAndServe(":8000", router))
}
