package main
import (
	"os"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
	"regexp"
	"bytes"
	"io/ioutil"
	"helper/helper/config"
	"helper/helper/logger"
)
var AppConfig *config.Config
//izinli ipleri kontrol et
func isAllowed(ip string) bool {
	for _, allowed := range AppConfig.AllowedIPs {
		if ip == allowed || "0.0.0.0" == allowed {
			return true
		}
	}
	return false
}


//smtp islemleri
func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(60 * time.Second))
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	clientIP := remoteAddr.IP.String()
	if !isAllowed(clientIP){
		logger.Log.Printf("Reddedildi: %s izinli değil",clientIP)
		conn.Write([]byte("554 Connection refused\r\n"))
		return
	}
	var from, to string
	var inDataMode bool
	var message strings.Builder

	buf := make([]byte,1024)
	conn.Write([]byte("220 sms-smtp.local ESMTP Service Ready\r\n"))

	for{
		n,err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Read error:",err)
			}
			return
		}
		data := string(buf[:n])
		data = strings.ReplaceAll(data,"\r\n","")

		if inDataMode {
			if data == "." {
				re := regexp.MustCompile(`Subject:\sAuthCode:\s([0-9]{6})`)
				matches := re.FindSubmatch([]byte(message.String()))
				if len(matches)>1 {
					conn.Write([]byte("250 Ok Message Accepted\r\n"))
					go sendSMS(to, string(matches[1]))
					inDataMode = false
					message.Reset()
				} else {
					logger.Log.Println("Onay kodu okunamadı !")
					conn.Write([]byte("552 Message not sent! \r\n"))
					return
				}
			}else{
				if message.Len() > 10*1024 {
					conn.Write([]byte("552 Message size exceeds fixed limit \r\n"))
					return
				}
				message.WriteString(data+"\n")
			}
			continue
		}
		switch {
			case strings.HasPrefix(strings.ToUpper(data), "HELO"),
			strings.HasPrefix(strings.ToUpper(data), "EHLO"):
			conn.Write([]byte("250 Hello\r\n"))

		case strings.HasPrefix(strings.ToUpper(data), "MAIL FROM:"):
			if from != "" {
				conn.Write([]byte("503 Sender already specified\r\n"))
				log.Println(data)
			}else {
				from = strings.TrimSpace(data[10:])
				from = strings.Trim(from,"<>")
				fromDomain := strings.Split(from,"@")
				if from == "" {
					conn.Write([]byte("501 Syntax error in MAIL FROM\r\n"))
				}else if len(fromDomain) < 2 {
					conn.Write([]byte("503 Bad sequence of commands\r\n"))
					return
				}else if fromDomain[1] != AppConfig.From{
					conn.Write([]byte("550 Sender domain not allowed\r\n"))
					return
				}else{
					conn.Write([]byte("250 OK\r\n"))
				}
			}
		case strings.HasPrefix(strings.ToUpper(data), "RCPT TO:"):
			if from == "" {
				log.Println(data)
				conn.Write([]byte("503 Need MAIL FROM before RCPT TO\r\n"))
			}else {
				to = strings.TrimSpace(data[8:])
				to = strings.Trim(to,"<>")
				to = strings.Split(to,"@")[0]
				re := regexp.MustCompile(`[0-9]{10}`)
				if !re.MatchString(to) {
					conn.Write([]byte("501 Syntax error in RCPT TO\r\n"))
					to = ""
				}
				if to == "" {
					conn.Write([]byte("501 Syntax error in RCPT TO\r\n"))
				}else{
					conn.Write([]byte("250 OK\r\n"))
				}
			}
		case strings.ToUpper(data) == "DATA":
			if from == "" || to == "" {
				conn.Write([]byte("503 Bad sequence of commands\r\n"))
			}else {
				inDataMode = true
				conn.Write([]byte("354 End data with <CR><LF>.<CR><LF>\r\n"))
			}
		case strings.ToUpper(data) == "QUIT":
			conn.Write([]byte("221 Bye\r\n"))
			return
		default:
			conn.Write([]byte("500 Unrecognized command\r\n"))
		}

	}
}

func sendSMS(phone string, message string){
	params := fmt.Sprintf(`data=<sms>
	<kno>%s</kno>
	<kulad>%s</kulad>
	<sifre>%s</sifre>
	<gonderen>%s</gonderen>
	<telmesajlar>
	<telmesaj>
	<tel>%s</tel><mesaj>Dogrulama kodunuz: %s</mesaj>
	</telmesaj>
	</telmesajlar>
	<tur>Normal</tur>
	</sms>`,AppConfig.MusteriNo,AppConfig.ApiID,AppConfig.ApiKey,AppConfig.Sender,phone,message)

	type RespJson struct {
		Code int `json:"code"`
		Status string `json:"status"`
		Description string `json:"description"`
	}
	req , err := http.NewRequest("POST","http://panel.vatansms.com/panel/smsgonderNNpost.php",bytes.NewBuffer([]byte(params)))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	status := strings.Split(string(body),":")
	if len(status)>3 {
		log.Println(fmt.Sprintf("Id Bilgisi %s Mesaj Gönderildi.",status[1]))
	}
}

func main(){
	cfg,err := config.LoadConfig("config.json")
	if err != nil {
		logger.Log.Fatal("Ayar dosyası yüklenirken hata oluştu")
	}
	AppConfig=cfg

	if err := logger.Init("app_errors.log"); err != nil {
		println("Log başlatılamadı:", err.Error())
		os.Exit(1)
	}
	listener, err := net.Listen("tcp", ":25")
	if err != nil {
		logger.Log.Fatal("Port dinleme hatası:", err)
	}
	log.Println("SMTP Server başlatıldı :25 portunda...")
	for{
		conn, err := listener.Accept()
		if err != nil {
			logger.Log.Println("Bağlantı hatası:", err)
			continue
		}
		go handleConnection(conn)
	}
}
