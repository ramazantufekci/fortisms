# FortiSMS

FortiSMS, **Fortigate SSL VPN** oturum açma sürecine **SMS tabanlı ikinci faktör doğrulama (2FA)** eklemek için geliştirilmiş bir Go uygulamasıdır.  
Uygulama Docker imajı olarak çalıştırılır ve gelen SMTP isteklerini dinleyerek SMS sağlayıcısı üzerinden OTP (one-time password) gönderir.

---

## 🚀 Özellikler
- Fortigate SSL VPN girişlerinde SMS doğrulama
- Basit ve hızlı Go tabanlı uygulama
- Docker imajı ile kolay kurulum ve dağıtım
- JSON tabanlı yapılandırma
- Log dosyası üzerinden hata takibi
- Vatansms alt yapısını kullanarak sms atmaktadır.

---

## 📦 Gereksinimler
- [Docker](https://docs.docker.com/get-docker/) (20.x veya üzeri önerilir)
- Fortigate cihazınızın **SMTP üzerinden OTP doğrulama** desteklemesi
- SMS sağlayıcısı için gerekli **API anahtarı ve kimlik bilgileri**

---

## ⚙️ Yapılandırma



Fortigate firewall cihazinda yapilandirmayi yapmak için [Fortigate SMS ile İki Faktörlü Kimlik Doğrulama Yapılandırması](https://www.ramazantufekci.com/fortigate-sms-ile-iki-faktorlu-kimlik-dogrulama-yapilandirmasi/) konulu makaleden yararlanabilirsiniz.

### `config.json`
Uygulama için gerekli ayarları içeren JSON dosyasıdır. Örnek:

```json
{
  "musterno": "123456",
  "api_id": "api-user",
  "api_key": "super-secret-key",
  "sender": "FORTISMS",
  "ListenPort": 25,
  "from": "example.com",
  "allowed_ips": ["1.1.1.1","0.0.0.0"]
}
````

> ⚠️ `from` alanı, uygulamanın kabul edeceği gönderici domainini belirtir.
> Bu domain dışında gelen SMTP istekleri reddedilir.

> ⚠️ `allowed_ips` alanı bağlantı yapacak olan smtp server ip adresidir. Eğer belirtmek istemezseniz "0.0.0.0" olarak yazmanız gerekmektedir.


---

## 🐳 Docker İmajı Oluşturma

Proje dizininde aşağıdaki komut ile Docker imajını oluşturun:

```bash
docker build -t fortisms:latest .
```

---

## ▶️ Uygulamayı Çalıştırma

Aşağıdaki komut ile uygulamayı başlatabilirsiniz:

```bash
docker run -d \
  --name fortisms \
  -e CONFIG_KEY=deneme \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/app_errors.log:/app/app_errors.log \
  -p 25:25 \
  ghcr.io/ramazantufekci/fortisms/fortisms:main
```

### Parametre Açıklamaları

* `-d` : Container’ı arka planda çalıştırır.
* `--name fortisms` : Container adı.
* `-e CONFIG_KEY=deneme` : Örnek ortam değişkeni (gerekirse değiştirin).
* `-v $(pwd)/config.json:/app/config.json` : Yerel `config.json` dosyasını container’a mount eder.
* `-v $(pwd)/app_errors.log:/app/app_errors.log` : Hata loglarını container dışına taşır.
* `-p 25:25` : SMTP servisi için 25. portu container’dan host’a yönlendirir.
* `fortisms:latest` : Kullanılacak Docker imajı.

---

## 📜 Loglama

Çalışma sırasında oluşan hatalar `app_errors.log` dosyasına yazılır.
Logları anlık izlemek için:

```bash
tail -f app_errors.log
```

Docker loglarını görmek için:

```bash
docker logs -f fortisms
```

---

## 🔧 Servisi Durdurma / Başlatma

```bash
docker stop fortisms
docker start fortisms
```

Servisi tamamen kaldırmak için:

```bash
docker rm -f fortisms
```

---

## 🛡️ Güvenlik Notları

* SMS sağlayıcısı API anahtarınızı `config.json` dışında kod tabanına koymayın(container ı çalıştırmadan önce config.json dosyasını oluşturmanız gerekmektedir. container çalıştıktan sonra CONFIG_KEY ile şifrelenir.).
* Port 25’in sadece Fortigate cihazınızdan erişilebilir olduğundan emin olun (firewall kuralı ekleyin).

---

## 🤝 Katkıda Bulunma

1. Bu repoyu forklayın.
2. Yeni bir dal (branch) oluşturun: `feature/xyz`
3. Değişikliklerinizi commit edin ve PR gönderin.

---

## 📄 Lisans

Bu proje MIT lisansı ile yayınlanmıştır.
Detaylar için [LICENSE](LICENSE) dosyasına bakın.
