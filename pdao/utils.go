package pdao

func toByte32(s []byte) [32]byte {
	a := [32]byte{}
	copy(a[:], s)
	return a
}
