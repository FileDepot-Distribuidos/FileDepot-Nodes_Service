package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	pb "filesystem/proto/filesystem"

	"google.golang.org/grpc"
)

const serverAddress = "localhost:50051"

func main() {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileSystemServiceClient(conn)
	time.Sleep(2 * time.Second)

	// Subir archivos
	uploadFile(client, "file1.txt", "Contenido del archivo 1")
	uploadFile(client, "file2.txt", "Contenido del archivo 2")
	uploadFile(client, "file3.txt", "Contenido del archivo 3")
}

func uploadFile(client pb.FileSystemServiceClient, filename, content string) {
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	res, err := client.UploadFile(context.Background(), &pb.UploadRequest{
		Filename:      filename,
		ContentBase64: encodedContent,
	})
	if err != nil {
		log.Printf("Error subiendo archivo %s: %v", filename, err)
	} else {
		fmt.Printf("✅ Subido: %s\n", filename)
		fmt.Printf("📁 Ruta: %s\n", res.FilePath)
		fmt.Printf("📄 Nombre: %s\n", res.FileName)
		fmt.Printf("📦 Tamaño: %d bytes\n", res.FileSize)
		fmt.Printf("🧾 Tipo MIME: %s\n", res.FileType)
		fmt.Printf("Node ID: %s\n", res.NodeId)
	}
	time.Sleep(1 * time.Second)
}
