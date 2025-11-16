# Vimedika - Apotek 🚀

[![Golang](https://img.shields.io/badge/Golang-1.25%2B-blue.svg)](https://golang.org/)
[![Postgres](https://img.shields.io/badge/PostgreSQL-17.4-yellow)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.4%2B-red)](https://www.redis.io)
[![JWT](https://img.shields.io/badge/TokenJWT-v5.3%2B-purple)](https://www.jwt.io/)
---

<div align="center">
    Layanan RESTful API performa tinggi untuk sistem manajemen apotek, dioptimalkan dengan caching Redis dan autentikasi JWT.
    <br />
    <br />
    <a href="https://vimedika.apidog.io">Lihat Dokumentasi API</a>
    ·
    <a href="[Link ke Issue Tracker]">Laporkan Bug</a>
    ·
    <a href="[Link ke Kontribusi Proyek]">Minta Fitur</a>
</div>

---

## 🧐 Tentang Proyek

Dalam repositori ini kita menerapkan `Golang` sebagai platform dasar bahasa pemrograman yang digunakan dalam pembuatan `API`.
Di dalam repositori ini juga kami terapkan framework `Scafold` yang kami buat sendiri serta dependensi `GORM` dan `JWT` untuk mempermudah dalam pengerjaan di ranah sekuritas maupun pengelolaan databasenya.

### 🛠️ Dibangun Dengan (The Tech Stack)

Proyek ini dikembangkan menggunakan teknologi-teknologi utama berikut:

* [![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
* [![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
* [![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
* **GORM:** ORM Golang yang luar biasa untuk interaksi database.
* **JWT (JSON Web Tokens):** Digunakan untuk autentikasi dan otorisasi API.
* **Scafold/Internal Framework:** Kumpulan *helper* dan *utility* khusus untuk mempercepat pengembangan.

---

## 🏁 Memulai (Getting Started)

Bagian ini memandu Anda untuk menyiapkan dan menjalankan proyek di lingkungan lokal Anda untuk tujuan pengembangan dan pengujian.

### ⚙️ Prerequisites (Prasyarat)

Pastikan Anda telah menginstal yang berikut ini:

* **Golang** (Versi 1.25 atau lebih tinggi)
* **PostgreSQL** (Database)
* **Redis** (Server Caching/Session)
* **Git**

### 📦 Installation (Instalasi)

1.  **Clone** repositori ini:
    ```bash
    git clone git@github.com:heru-oktafian/api-retail.git
    cd api-retail
    ```

2.  **Siapkan Database:**
    * Buat database PostgreSQL baru.
    * Konfigurasi koneksi database Anda di file `.env`.

3.  **Siapkan Environment (Lingkungan):**
    * Duplikasi file `.env.example` dan ganti namanya menjadi `.env`.
    * Isi variabel-variabel yang diperlukan (`DB_HOST`, `DB_USER`, `REDIS_HOST`, `JWT_SECRET`, dll.).

4.  **Jalankan Migrasi Database (Jika Menggunakan GORM Migrations):**
    ```bash
    go run [path/ke/file/migrasi/utama].go
    ```
    *[Sesuaikan perintah migrasi Anda]*

5.  **Jalankan Proyek:**
    ```bash
    go run main.go
    # Atau gunakan: go build && ./[nama executable]
    ```

Proyek akan berjalan di `http://localhost:9002`.

---

## 🤸 Penggunaan API (Usage)

API ini dirancang untuk mengelola inventaris produk, pembelian, penjualan, retur, opaname, data pelanggan, dll.

### Contoh Autentikasi

Semua *endpoint* yang aman memerlukan token **Bearer JWT** di *header*.

| Header | Nilai |
| :--- | :--- |
| `Authorization` | `Bearer <your_jwt_token>` |

### Endpoint Utama

| Kategori | Deskripsi |
| :--- | :--- |
| `/api/auth` | Pendaftaran & *Login* Pengguna. |
| `/api/products` | Manajemen Inventaris & Produk. |
| `/api/sales` | Pembuatan & Pelacakan Pesanan. |
| `/api/users` | Pengelolaan Data Pengguna. |

**Lihat dokumentasi lengkap di [vimedika.apidog.io](https://vimedika.apidog.io)**

---

## 🛣️ Roadmap (Rencana Pengembangan)

* Penambahan Laporan Stock Opaname.
* Penambahan Laporan Aset.
* Penambahan Fitur Penandaan Pembayaran Tagihan.
* Optimasi pencarian produk dengan sistem caching yang lebih baik.

---

## 🤝 Kontribusi (Contributing)

Kontribusi adalah hal yang membuat komunitas *open source* menjadi tempat yang luar biasa untuk belajar, menginspirasi, dan berkreasi. Setiap kontribusi yang Anda berikan sangat **dihargai**.

Jika Anda memiliki saran yang akan membuat ini lebih baik, silakan *fork* repo dan buat *Pull Request*. Anda juga dapat membuka *issue* dengan tag "enhancement".

1.  *Fork* Proyek.
2.  Buat *Branch* Fitur Anda (`git checkout -b feature/AmazingFeature`).
3.  *Commit* Perubahan Anda (`git commit -m 'Add some AmazingFeature'`).
4.  *Push* ke *Branch* (`git push origin feature/AmazingFeature`).
5.  Buka *Pull Request*.

---

## 📄 Lisensi (License)

Karya ini (termasuk semua kode dan konten di repositori ini) dilindungi oleh hak cipta. Penggunaan, penyalinan, atau modifikasi dalam bentuk apa pun dilarang tanpa izin tertulis dari saya.

Untuk meminta izin, silakan hubungi saya di heru@heruoktafian.com.

---

## ✉️ Kontak (Contact)

Heru Oktafian, ST., CTT - [@heru-oktafian](https://x.com/HeruOktafianST) - [heru@heruoktafian.com](mailto:heru@heruoktafian.com)

Tautan Proyek: [https://github.com/heru-oktafian/api-apotek](https://github.com/heru-oktafian/api-apotek)