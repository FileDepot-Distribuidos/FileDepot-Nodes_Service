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

	// Crear directorios
	createDirectory(client, "newdir")
	createSubdirectory(client, "newdir", "subdir")

	// Listar archivos y carpetas
	listAll(client, "newdir")

	// Operaciones sobre archivos
	deleteFile(client, "file1.txt")
	renameAndMoveFile(client, "file2.txt", "newdir/renamed.txt")
	moveFile(client, "file3.txt", "newdir/subdir/file3.txt")

	// Listar nuevamente
	listAll(client, "newdir")
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
		fmt.Printf("Subido: %s -> %s\nRuta: %s\n", filename, res.Message, res.FilePath)
	}
	time.Sleep(1 * time.Second)
}

func createDirectory(client pb.FileSystemServiceClient, path string) {
	res, err := client.CreateDirectory(context.Background(), &pb.DirectoryRequest{Path: path})
	if err != nil {
		log.Printf("Error creando directorio %s: %v", path, err)
	} else {
		fmt.Println(res.Message)
	}
	time.Sleep(1 * time.Second)
}

func createSubdirectory(client pb.FileSystemServiceClient, parent, subdir string) {
	res, err := client.CreateSubdirectory(context.Background(), &pb.SubdirectoryRequest{
		ParentDirectory:  parent,
		SubdirectoryName: subdir,
	})
	if err != nil {
		log.Printf("Error creando subdirectorio %s: %v", subdir, err)
	} else {
		fmt.Println(res.Message)
	}
	time.Sleep(1 * time.Second)
}

func deleteFile(client pb.FileSystemServiceClient, path string) {
	res, err := client.DeleteFile(context.Background(), &pb.DeleteRequest{Path: path})
	if err != nil {
		log.Printf("Error eliminando archivo %s: %v", path, err)
	} else {
		fmt.Println(res.Message)
	}
	time.Sleep(1 * time.Second)
}

func renameAndMoveFile(client pb.FileSystemServiceClient, oldPath, newPath string) {
	res, err := client.RenameFile(context.Background(), &pb.RenameRequest{
		OldName: oldPath,
		NewName: newPath,
	})
	if err != nil {
		log.Printf("Error renombrando/moviendo archivo %s: %v", oldPath, err)
	} else {
		fmt.Println(res.Message)
	}
	time.Sleep(1 * time.Second)
}

func moveFile(client pb.FileSystemServiceClient, source, destination string) {
	res, err := client.MoveFile(context.Background(), &pb.MoveRequest{
		SourcePath:      source,
		DestinationPath: destination,
	})
	if err != nil {
		log.Printf("Error moviendo archivo %s: %v", source, err)
	} else {
		fmt.Println(res.Message)
	}
	time.Sleep(1 * time.Second)
}

func listAll(client pb.FileSystemServiceClient, path string) {
	res, err := client.ListFiles(context.Background(), &pb.DirectoryRequest{Path: path})
	if err != nil {
		log.Printf("Error listando contenido en %s: %v", path, err)
		return
	}

	fmt.Printf("\n📂 Contenido de '%s':\n", path)
	if len(res.Files) > 0 {
		for _, item := range res.Files {
			fmt.Println("", item)
		}
	} else {
		fmt.Println("   (Vacío)")
	}
	fmt.Println()
	time.Sleep(1 * time.Second)
}
