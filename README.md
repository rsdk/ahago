ahago
=====

Library to use the AVM Home Automation Control in Go via AVM HTTP Interface


=====

Example code:

package main
import (
	"github.com/rsdk/ahago"
)

func main() {
	conn := ahago.Connect("username", "password")	//Verbinden mit Fritzbox
	conn.GetStatus()		//Abfragen des Status aller mit der Fritzbox registrierten Geräte
	conn.Close()			//Verbindung geschlossen
}
