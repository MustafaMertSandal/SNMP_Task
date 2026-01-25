# SNMP Task (GNS3 + SNMP + TimescaleDB + Web API)

Bu proje, GNS3 sanallaştırma ortamında çalışan ağ cihazlarından SNMP ile veri toplayan,
verileri TimescaleDB’de saklayan ve Web API üzerinden görüntüleyen Golang tabanlı bir uygulamadır.

## Özellikler
- SNMP polling (v2c) ile cihazlardan sistem bilgileri + interface metrikleri + ip route bilgileri toplama
- TimescaleDB’ye yazma (batch insert) ve 1 dakikalık downsample (continuous aggregate)
- Otomatik veri retention (raw tablolar için 10 dakikadan eskiler ve 1 dakikalık (1m) aggregate tabloları için 20 dakikadan eskiler silinir).
- REST API + basit frontend (cihaz listesi + device detayları)
- OpenAPI dokümantasyonu
- Export edilmiş 2 router bulunan GNS3 projesi
![GNS3 Projesi](<images/Screenshot from 2026-01-25 16-13-46.png>)

---

## Gereksinimler
- **GNS3** 
- **Docker**
- **Docker Compose**

---

## Kurulum ve Çalıştırma

### 1) Projeyi indir
- ```bash
  git clone <https://github.com/MustafaMertSandal/SNMP_Task.git>
  cd <REPO_FOLDER>


### 2) GNS3 Topolojisini Ayarla
- GNS3 uygulaması -> File -> Import portable project -> Projenin gns3 klasörü altındaki gns3_project.gns3_project dosyasını import et.
- Projeyi çalıştır.
- Bu topolojide bağlantı kurabilmek için tap0 kullanılmıştır (Linux). Kendine göre dizayn edebilirsin: 
  ```bash
  sudo ip tuntap add dev tap0 mode tap user $USER
  sudo ip link set tap0 up
  sudo ip addr add 10.10.10.2/24 dev tap0
  sudo ip route add 10.10.20.0/24 via 10.10.10.1 dev tap0
- Bağlantıyı terminalinden test et.
  ```bash
  snmpget -v2c -c public 10.10.10.1 1.3.6.1.2.1.1.1.0
  

### 3) Konfigürasyon

`config.yaml`
- Konfigürasyon dosyası paylaşılan topolojiye uygun olarak ayarlanmıştır.
> Not: SNMP community ve hedef IP/port değerleri GNS3 topolojinize göre ayarlanmalıdır.

- `targets`: SNMP ile okunacak cihazlar
- `oids`: toplanacak OID listesi
- `database`: TimescaleDB bağlantısı
- `web`: web server adresi

### 4) Çalıştırma
- ```bash
  docker compose down -v
  docker compose up --build -d
- Logları izle:
  - ```bash
    docker logs -f timescaledb
  - Ayrı terminal:
    ```bash
    docker logs -f snmp-task


### 5) Açılan Servisler
- snmp-task logunda localhost linkinden web'de projeyi görebilirsin.
![Web-ui](<images/Screenshot from 2026-01-25 16-39-32.png>)
- Burada gösterilen veriler downsample edilmiş verilerdir. Bundan dolayı veritabanı boş ise verilerin gelmesi için yaklaşık 3 dakika bekle.
