package s5

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mvisonneau/s5/cipher"
	cipherPGP "github.com/mvisonneau/s5/cipher/pgp"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/stretchr/testify/assert"
)

const testPGPPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBF5Slx8BEADISfn9MzCsrAvLonPhwAYVlEFWWqk3Z6gZQekwdsTp6tyQWYWD
J1I7yzGr05tUKUAvrNaCfWH6syh8sHVp+7iYGJjDzbPIl+yn+8grCWhEa0+23iPw
EBX+Q5iTlp2ZmwIpbD7/XL0Y/dsJZID80k4lhPxIPbH2a+cPOIstmtGhXKJPBItA
I2G/fSwWJpivGLeYrIXJxHAkCG9GTnmo2u+s+Kl9A5m7zOPSgnV+Wcl+ofX6RMDZ
qScTsaq+NfyEH0rC7lSvn+BxrBP8rjP4abtB1EpiAJaZBnu7Vj8d7YrtrTu5eWnX
cO4C1sd84AB+Gz4M+6lBaWnTvSOWfWJeM6O4qa++mjnO4bN4cRrneiEz1/vuzL5o
NF5ky04Ro16MmDN3XmzXidCUJaEBjZ3vfxTqZ9RQqWX3Kzpl7bN/OzXXBcd4EFgv
FULhiwquYAG++0PrPLOMEAZhaPpNKHAiz1Hxiztai1WT+CLBQWz+ERLpk7ZoPRTd
ZqMopZcOX821d28AxDi80X/Imhd37hkSEwfMjW3ZKmNStSzXoTUriE4GZIGIJa33
xyhgRtO/eRB+I3XnEU2BEdAqiI8nDRA2p4myZE9EtS1hqsCdYnTZrbxrMbuxBYvO
rLfbAS/dfvqPkB8K0VIZuH4m8vqeZ+vwhOkYVj7pU6eoTTffh6x4zNGebQARAQAB
tBVGb28gQmFyIDxmb29AYmFyLmNvbT6JAk4EEwEIADgWIQT92jKxeZxZcBEPNm+N
0lu3UxLgYQUCXlKXHwIbLwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRCN0lu3
UxLgYSOvEACUSCl8XAVaK31XiUk5qQqef9BXhM6Hko5mQYQUZ2XAGzOJ7HJiaja0
RW0oSpqvdIwEdHkxWPxkmTY7QpWN9pm9U+XXOJlL0BcNvqVFOmIuTawqlRWSOq1z
QY/Y6951vovJVfVCRFSUgfsvXFkTBT+ke5wytyf8Y36agldNyAA3KJ4ykEi8pNTP
fXrABabJbFG+MqFkDR/9bwzhj5AXpJXrZxtojYP+QMI1H2xSkT0rg0gPJaa87Fn7
g6ImLSHRZ9L/35uS5ieLPC9g94M3o9NCdLJagfKnq8OgEzQfrroF9qndOl3rvtz2
hD+lB7xS2StUAOBgXrVCBS/hVNglHGAqZgFRsRZTmkMIzjFw27fVi0UuPB96M8dr
HzED29MD7r1OeFlpKLrYDKkWkn9RJw04cA7uMWqF2SREbXAy6dAD6+Rj2riKwaoo
HJ9sB7ZnOJ7eFtOroiLe/RZfy30khM5pfs6dsmRH9od+cb6d6uEHgC74LZdZKI7R
L/lhQ0S+LoqCqjqt88FSdY91P4p6mSUuOAwx0uwH++Q2FwG/0E8f8Xhd0ktgC8S7
iTz7MehdL+CFeb0eBQsylgh4naslFUky3CvqKtm5fwYO1I7fsJUM9UjUZsYu1PSw
+1cywObMR/cznH3QV3cLl40WjVUoVG+jAINcqnZNlU6Xjz6nIrxwHg==
=JdiX
-----END PGP PUBLIC KEY BLOCK-----`

const testPGPPrivateKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQcYBF5Slx8BEADISfn9MzCsrAvLonPhwAYVlEFWWqk3Z6gZQekwdsTp6tyQWYWD
J1I7yzGr05tUKUAvrNaCfWH6syh8sHVp+7iYGJjDzbPIl+yn+8grCWhEa0+23iPw
EBX+Q5iTlp2ZmwIpbD7/XL0Y/dsJZID80k4lhPxIPbH2a+cPOIstmtGhXKJPBItA
I2G/fSwWJpivGLeYrIXJxHAkCG9GTnmo2u+s+Kl9A5m7zOPSgnV+Wcl+ofX6RMDZ
qScTsaq+NfyEH0rC7lSvn+BxrBP8rjP4abtB1EpiAJaZBnu7Vj8d7YrtrTu5eWnX
cO4C1sd84AB+Gz4M+6lBaWnTvSOWfWJeM6O4qa++mjnO4bN4cRrneiEz1/vuzL5o
NF5ky04Ro16MmDN3XmzXidCUJaEBjZ3vfxTqZ9RQqWX3Kzpl7bN/OzXXBcd4EFgv
FULhiwquYAG++0PrPLOMEAZhaPpNKHAiz1Hxiztai1WT+CLBQWz+ERLpk7ZoPRTd
ZqMopZcOX821d28AxDi80X/Imhd37hkSEwfMjW3ZKmNStSzXoTUriE4GZIGIJa33
xyhgRtO/eRB+I3XnEU2BEdAqiI8nDRA2p4myZE9EtS1hqsCdYnTZrbxrMbuxBYvO
rLfbAS/dfvqPkB8K0VIZuH4m8vqeZ+vwhOkYVj7pU6eoTTffh6x4zNGebQARAQAB
AA/8Cj5WVClpEa7rteqrq8REmDGKTrgf5wVP88xXxWXECzg60M5ZslJ+TzArBmAY
Yp6IrNq76Dww/z+bqeNKFGiOXiEFg+yIusQ7+bKgrMRnZvbO3EoFY0+MGgJen4KL
5U5BnXK7nUqiu50U6PV3iPamKB8Vpzz+70NBGjMvH976KufwJfbUISyhR2GuGCS5
mSlV3BYYWMn8kXgCkOfBXlRymGsc5/4ZWVSD4ye1Vwf3WTJy84c1DpMrIqoeH52A
bUilOM3Np2NkeQvVXdsrx0Z8Q3dv2Nrk2OC1aIH86mZUVtsne2q6C6dRiCc9JQxw
mgKBeVKDRPHD1G0PXX2Xwu+eYOOR+Lty6wzgI/yQgmLL+vZQr3vfcPUJLCVVwWAu
284YREQwaYWjMuTort0bl7SbiseAZHiigWXOQ88KZr1IIVyAGYi6Gv0Fpo2at8D8
33srlt6wnG1raJzDBKboGsU+2IByxg1pspPSGvurNSBKHnaccSGT4+fHt9bxO32Q
zZW5tgfnWqXZ4A4y9luoDeQtk3iK6MB1Tkhl0rJZI3HO/VjniYY1uFqf5gaZleDT
y8DOlDzfk1BXvsrkLYbZN8Wm5nnsDD4Apv2xmdCiUV6HiquYd7jnR2XjAS/bJVKE
JBYFIoIBSGkrc8ktUqz2xgX6j3fm1oGWzuHk/leXD5k3JzUIANEKETfpTciSTwnE
7oJnacX0rtXhBmg0eyquTTOXAvq6bup7e00J0nPI2QN00OsvjQg+ntR2fDhJdASW
hOp/MEB5mQ9DQM8Huk65GwFxxCjspb62Mx3//Jj33RdJsSyzWPyfa9emH+0toUHI
PSSAJ3GIBzhCwIQmAzE4+hZeju9u9hv0pMcLkSRIBL95QALAstOthuVb7PswMrW2
bMN0FF0X5tEUlABpzqI8tXFgd7VL0gWhYHMdEyh4sTTsDanL62oRFuyIBDK1U6Sy
4AUT58urJxUMZSAhCec9b2ejV2IQg3+v4idxWdvzK4eOObQvIoc10+1HVRBxfXzK
Q8LvfeMIAPVIrEhbli28mWUJhefLzVjq7exoKBDknBDvqhN3vWrgEYHGu8f01RYR
XeHKiIcXcDwxoLS8nFzupw2e3ab/rgJHYF6Me3hco4by3xylqmDS/7k0we5DtOs6
WDZs5z6A46yWwvwcF63JZCyw/VSv0NGI0Eh9MNBPMkvprarDADX2zald8YvnQztZ
ol3HR+U2GGM3wJdEYZQ31D7GFO0VhdOH7zFuj3PuvlyYpwOaeRozaXvedfRULNG1
TrzBuVp4+4IAPWwCRTskTQHr69DS8sLOX0cMQj+OIn/LBW9ScCbgMEjOIe7CdZM5
kC7ohb08fBMUQ7ph0RFu0l0Uhm9sI28IAI0n/YfGPI1ELrXHC00x87ETobczIr4W
APZOl4CpJujWFUHGRtRticWUs0v5nmPODf1ieshuA8PwNJCBYTBN/33x09ygAdym
huJa8u4v4I9iHcFl9l6MwcMU/WzrvPYv0EWO2bbRmW4UUbHXnXQFed/v6psV9dgI
OlZbbE+5MUyARgG00utw9cDmwQT0pOK1H5ksBPKsMLo8K/o31vKuLBAYIISZ8srg
gu7hb/vx8Kl6+226yEbGADli5/aXx2kGJo36j5Udx8qWXwiJQl0zjLTAGRhFxVDE
eHAmlXT0rUpRXI7/7jKwl621G+E57GF/x0FG6bkJvJIzjPJCOmXfBN6LxrQVRm9v
IEJhciA8Zm9vQGJhci5jb20+iQJOBBMBCAA4FiEE/doysXmcWXARDzZvjdJbt1MS
4GEFAl5Slx8CGy8FCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQjdJbt1MS4GEj
rxAAlEgpfFwFWit9V4lJOakKnn/QV4TOh5KOZkGEFGdlwBsziexyYmo2tEVtKEqa
r3SMBHR5MVj8ZJk2O0KVjfaZvVPl1ziZS9AXDb6lRTpiLk2sKpUVkjqtc0GP2Ove
db6LyVX1QkRUlIH7L1xZEwU/pHucMrcn/GN+moJXTcgANyieMpBIvKTUz316wAWm
yWxRvjKhZA0f/W8M4Y+QF6SV62cbaI2D/kDCNR9sUpE9K4NIDyWmvOxZ+4OiJi0h
0WfS/9+bkuYnizwvYPeDN6PTQnSyWoHyp6vDoBM0H666Bfap3Tpd677c9oQ/pQe8
UtkrVADgYF61QgUv4VTYJRxgKmYBUbEWU5pDCM4xcNu31YtFLjwfejPHax8xA9vT
A+69TnhZaSi62AypFpJ/UScNOHAO7jFqhdkkRG1wMunQA+vkY9q4isGqKByfbAe2
Zzie3hbTq6Ii3v0WX8t9JITOaX7OnbJkR/aHfnG+nerhB4Au+C2XWSiO0S/5YUNE
vi6Kgqo6rfPBUnWPdT+KepklLjgMMdLsB/vkNhcBv9BPH/F4XdJLYAvEu4k8+zHo
XS/ghXm9HgULMpYIeJ2rJRVJMtwr6irZuX8GDtSO37CVDPVI1GbGLtT0sPtXMsDm
zEf3M5x90Fd3C5eNFo1VKFRvowCDXKp2TZVOl48+pyK8cB4=
=ueeG
-----END PGP PRIVATE KEY BLOCK-----`

func createTestPGPKeys(publicKey, privateKey string) (string, string, error) {
	tmpFilePublic, err := ioutil.TempFile(os.TempDir(), "tfcw-test-pgp-pub-")

	if _, err = tmpFilePublic.Write([]byte(publicKey)); err != nil {
		return "", "", fmt.Errorf("Failed to write to temporary file : %s", err.Error())
	}

	if err = tmpFilePublic.Close(); err != nil {
		return "", "", fmt.Errorf("Failed to close temporary file : %s", err.Error())
	}

	tmpFilePrivate, err := ioutil.TempFile(os.TempDir(), "tfcw-test-pgp-pri-")
	if _, err = tmpFilePrivate.Write([]byte(privateKey)); err != nil {
		return "", "", fmt.Errorf("Failed to write to temporary file : %s", err.Error())
	}

	if err = tmpFilePrivate.Close(); err != nil {
		return "", "", fmt.Errorf("Failed to close temporary file : %s", err.Error())
	}

	return tmpFilePublic.Name(), tmpFilePrivate.Name(), nil
}

func TestGetCipherEnginePGP(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypePGP
	publicKeyPath, privateKeyPath, err := createTestPGPKeys(testPGPPublicKey, testPGPPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(publicKeyPath)
	defer os.Remove(privateKeyPath)

	// expected engine
	expectedEngine, err := cipher.NewPGPClient(publicKeyPath, privateKeyPath)
	assert.Equal(t, err, nil)

	// all defined in client, empty variable config (default settings)
	v := &schemas.S5{}
	c := &Client{
		CipherEngineType: &cipherEngineType,
		CipherEnginePGP: &schemas.S5CipherEnginePGP{
			PublicKeyPath:  &publicKeyPath,
			PrivateKeyPath: &privateKeyPath,
		},
	}

	cipherEngine, err := c.getCipherEngine(v)
	assert.Equal(t, err, nil)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrimaryKey, *expectedEngine.Entity.PrimaryKey)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrivateKey, *expectedEngine.Entity.PrivateKey)

	// all defined in variable, empty client config
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEnginePGP: &schemas.S5CipherEnginePGP{
			PublicKeyPath:  &publicKeyPath,
			PrivateKeyPath: &privateKeyPath,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Equal(t, err, nil)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrimaryKey, *expectedEngine.Entity.PrimaryKey)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrivateKey, *expectedEngine.Entity.PrivateKey)

	// key defined in environment variable
	os.Setenv("S5_PGP_PUBLIC_KEY_PATH", publicKeyPath)
	os.Setenv("S5_PGP_PRIVATE_KEY_PATH", privateKeyPath)
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Equal(t, err, nil)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrimaryKey, *expectedEngine.Entity.PrimaryKey)
	assert.Equal(t, *cipherEngine.(*cipherPGP.Client).Entity.PrivateKey, *expectedEngine.Entity.PrivateKey)
}
