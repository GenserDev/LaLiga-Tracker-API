package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hola LaLiga Tracker!")
    })

    fmt.Println("Servidor iniciado en :8080")
    http.ListenAndServe("0.0.0.0:8080", nil)
}