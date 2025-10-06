package config
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	ApiID string `json:"api_id"`
	ApiKey string `json:"api_key"`
	Sender string `json:"sender"`
	Debug bool `json:"debug"`
	Timeout int `json:"timeout"`
	AllowedIPs []string `json:"allowed_ips"`
	From string `json:"from"`
	MusteriNo string `json:"musterino"`
}

func encrypt(plain []byte,key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]

	if _, err:= rand.Read(iv); err != nil {
	    return "",err
        }

        stream := cipher.NewCFBEncrypter(block, iv)
        stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)

        return base64.StdEncoding.EncodeToString(ciphertext), nil
}


func decrypt(cipherStr string, key []byte) ([]byte, error) {
	ciphertext, err:=base64.StdEncoding.DecodeString(cipherStr)
	if err !=nil {
		return nil, err
	}

	block, err:= aes.NewCipher(key)
	if err!=nil {
		return nil,err
	}

	if len(ciphertext)< aes.BlockSize{
		return nil, errors.New("ciphertext çok kısa")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block,iv)
	stream.XORKeyStream(ciphertext, ciphertext)

        return ciphertext, nil
}




func LoadConfig(path string) (*Config, error ){
	keyStr := os.Getenv("CONFIG_KEY")
	if keyStr == "" {
		return nil, errors.New("CONFIG_KEY environment variable yok, uygulama durduruldu")
	}

	key := []byte(keyStr)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("CONFIG_KEY uzunluğu geçersiz: %d byte. 16, 24 veya 32 byte olmalı", len(key))
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config dosyası okunamadı: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err == nil {
		fmt.Println("Config şifrelenmemiş, şifreleniyor...")

		enc, err := encrypt(content, key)
		if err != nil {
			return nil, fmt.Errorf("config şifrelenemedi: %v", err)
		}

		if err := ioutil.WriteFile(path, []byte(enc),0600); err != nil {
			return nil, fmt.Errorf("şifreli config kaydedilmedi :%", err)
		}
                    return &cfg, nil
	    }
	    dec, err :=decrypt(string(content),key)
	    if err!= nil{
			  return nil, fmt.Errorf("config çözülemedi: %v", err)
		  }
		  if err := json.Unmarshal(dec,&cfg); err !=nil {
			  return nil,fmt.Errorf("config JSON parse hatası :%v", err)
		  }
		  
		  return &cfg, nil
		  
	}

