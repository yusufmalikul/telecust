Create a telegram bot customer service that can answer user question using AI.

Flow:
1. User init chat e.g. Halo kak
2. bot answer "Apa yang bisa saya bantu, kak?"
3. Ingin tanya harga keripik kentang
4. bot answer "Harga keripik kentang satu bungkusnya Rp5ribu, jika pesan banyak ada potongan".

Bot answer based on knowledge base.
Example knowledge base:
"Harga kentang Rp5ribu perbungkus.
Pesan diatas 10 harga 4rb.
Jika pesan 10 Rp40ribu.
Jika pesan 20 Rp80ribu.
Jika pesan di atas 100 bungkus harga Rp3ribu".

Dashboard admin menunjukkan semua chat customer yang dilayani bot.
Ada tombol take over untuk mensetop bot.
Ada setting juga untuk mengatur knowledge base.

Admin can reply to customer chat via web too as bot.


Tech:
Golang, SQLite. (frontend up to you)