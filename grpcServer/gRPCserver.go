package main

import (
	pb "filesystem/proto/filesystem"
	"filesystem/server"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env, usando valores por defecto.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error al escuchar en el puerto %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	server := server.NewServer()

	pb.RegisterFileSystemServiceServer(grpcServer, server)

	log.Println("Servidor gRPC corriendo en el puerto 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
	}
}
