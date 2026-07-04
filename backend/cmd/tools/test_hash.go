//go:build ignore

package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	hash := "$2a$14$gG/9QB92MPRGAU7MsWidiOU7oXAfuO09eMlelbH.kmcm7neP8QPZe"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123"))
	fmt.Println(err == nil)
}
