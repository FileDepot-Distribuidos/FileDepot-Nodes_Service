package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "filesystem/proto/filesystem"

	"google.golang.org/grpc"
)

const serverAddr = "localhost:50051"

func main() {
	// Conectar con el servidor gRPC
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error conectando al servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileSystemServiceClient(conn)

	// Ejecutar pruebas con pausas entre ellas
	testUploadFile(client, "file1.txt")
	time.Sleep(1 * time.Second)

	testUploadFile(client, "file2.txt")
	time.Sleep(1 * time.Second)

	testRenameFile(client, "file1.txt", "file1_renamed.txt")
	time.Sleep(1 * time.Second)

	testDeleteFile(client, "file2.txt")
	time.Sleep(1 * time.Second)

	testCreateDirectory(client)
	time.Sleep(1 * time.Second)

	testMoveFile(client)
	time.Sleep(1 * time.Second)

	testListFiles(client)
	time.Sleep(1 * time.Second)
}

// Subir un archivo
func testUploadFile(client pb.FileSystemServiceClient, filename string) {
	fmt.Printf("=== Prueba: Subir archivo %s ===\n", filename)

	fileContent := []byte("Contenido de prueba")
	req := &pb.UploadRequest{
		Filename:  filename,
		Content:   fileContent,
		Directory: "", // Opcional: especificar un subdirectorio
	}

	res, err := client.UploadFile(context.Background(), req)
	if err != nil {
		log.Printf("Error al subir archivo %s: %v\n", filename, err)
	} else {
		fmt.Println(res.Message)
	}
}

// Crear un directorio
func testCreateDirectory(client pb.FileSystemServiceClient) {
	fmt.Println("=== Prueba: Crear directorio ===")

	req := &pb.DirectoryRequest{Path: "test_dir"}
	res, err := client.CreateDirectory(context.Background(), req)
	if err != nil {
		log.Printf("Error al crear directorio: %v\n", err)
	} else {
		fmt.Println(res.Message)
	}
}

// Mover un archivo
func testMoveFile(client pb.FileSystemServiceClient) {
	fmt.Println("=== Prueba: Mover archivo ===")

	req := &pb.MoveRequest{
		SourcePath:      "storage/file1_renamed.txt",
		DestinationPath: "storage/test_dir/file1_renamed.txt",
	}

	res, err := client.MoveFile(context.Background(), req)
	if err != nil {
		log.Printf("Error al mover archivo: %v\n", err)
	} else {
		fmt.Println(res.Message)
	}
}

// Renombrar un archivo
func testRenameFile(client pb.FileSystemServiceClient, oldName, newName string) {
	fmt.Printf("=== Prueba: Renombrar archivo %s a %s ===\n", oldName, newName)

	req := &pb.RenameRequest{
		OldName: oldName,
		NewName: newName,
	}

	res, err := client.RenameFile(context.Background(), req)
	if err != nil {
		log.Printf("Error al renombrar archivo: %v\n", err)
	} else {
		fmt.Println(res.Message)
	}
}

// Eliminar un archivo
func testDeleteFile(client pb.FileSystemServiceClient, filename string) {
	fmt.Printf("=== Prueba: Eliminar archivo %s ===\n", filename)

	req := &pb.DeleteRequest{Path: filename}
	res, err := client.DeleteFile(context.Background(), req)
	if err != nil {
		log.Printf("Error al eliminar archivo %s: %v\n", filename, err)
	} else {
		fmt.Println(res.Message)
	}
}

// Listar archivos en un directorio
func testListFiles(client pb.FileSystemServiceClient) {
	fmt.Println("=== Prueba: Listar archivos ===")

	req := &pb.DirectoryRequest{Path: "test_dir"}
	res, err := client.ListFiles(context.Background(), req)
	if err != nil {
		log.Printf("Error al listar archivos: %v\n", err)
	} else {
		fmt.Println("Archivos en test_dir:", res.Files)
	}
}
