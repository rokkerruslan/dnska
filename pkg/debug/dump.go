package debug

import (
	"os"
	"path"

	"github.com/google/uuid"
)

func DumpMalformedPacket(buf []byte) {
	id, _ := uuid.NewRandom()
	_ = os.WriteFile(path.Join("./dumps", id.String()), buf, os.FileMode(0766))
}
