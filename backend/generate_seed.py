import random

categories = [
    (10, 'Milliy Taomlar', 'https://images.unsplash.com/photo-1529042410759-befb1204b468?auto=format&fit=crop&q=80&w=400'),
    (11, 'Fast Food', 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400'),
    (12, 'Ichimliklar', 'https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400'),
    (13, 'Salatlar', 'https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400'),
    (14, 'Shirinliklar', 'https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400'),
]

# (category_id, name, desc, price, unit, min_qty, step, container)
milliy = [
    ("Osh (Palov)", 25000, "pors", 1, 1, True),
    ("Qozon kabob", 45000, "pors", 1, 1, True),
    ("Lag'mon", 22000, "pors", 1, 1, True),
    ("Manti", 4000, "dona", 4, 1, True),
    ("Somsa", 7000, "dona", 1, 1, False),
    ("Sho'rva", 20000, "pors", 1, 1, True),
    ("Shashlik (Qiyma)", 12000, "dona", 2, 1, True),
    ("Shashlik (Jaz)", 15000, "dona", 2, 1, True),
    ("Norin", 30000, "pors", 1, 1, True),
    ("Hasip", 25000, "pors", 1, 1, True),
    ("Qovurdoq", 40000, "pors", 1, 1, True),
    ("Tuxum barak", 18000, "pors", 1, 1, True),
    ("Dimlama", 35000, "pors", 1, 1, True),
    ("Beshbarmoq", 50000, "pors", 1, 1, True),
    ("Moshxo'rda", 18000, "pors", 1, 1, True),
    ("Chuchvara", 20000, "pors", 1, 1, True),
    ("Kavob", 35000, "pors", 1, 1, True),
    ("Go'shtli non", 15000, "dona", 1, 1, False),
    ("Qo'y go'shti (Xom)", 95000, "kg", 1, 0.5, False),
    ("Mol go'shti (Xom)", 85000, "kg", 1, 0.5, False),
]

fast_food = [
    ("Burger", 25000, "dona", 1, 1, True),
    ("Cheeseburger", 28000, "dona", 1, 1, True),
    ("Hot-dog", 15000, "dona", 1, 1, True),
    ("Klab Sendvich", 30000, "dona", 1, 1, True),
    ("KFC Tovuq (Qanotcha)", 35000, "pors", 1, 1, True),
    ("KFC Tovuq (Strips)", 32000, "pors", 1, 1, True),
    ("Fri kartoshkasi", 12000, "pors", 1, 1, True),
    ("Nonli hot-dog", 14000, "dona", 1, 1, True),
    ("Pitsa (Margarita)", 55000, "dona", 1, 1, True),
    ("Pitsa (Go'shtli)", 75000, "dona", 1, 1, True),
    ("Pitsa (Qo'ziqorinli)", 65000, "dona", 1, 1, True),
    ("Pitsa (Arlash)", 70000, "dona", 1, 1, True),
    ("Donar Kebab", 28000, "dona", 1, 1, True),
    ("Lavash (Go'shtli)", 26000, "dona", 1, 1, False),
    ("Lavash (Tovuqli)", 24000, "dona", 1, 1, False),
    ("Shaurma", 25000, "dona", 1, 1, False),
    ("Gamburger mini", 15000, "dona", 1, 1, True),
    ("Katta Combo", 60000, "pors", 1, 1, True),
    ("Kichik Combo", 45000, "pors", 1, 1, True),
    ("Qovurilgan Tovuq", 40000, "pors", 1, 1, True),
]

ichimliklar = [
    ("Coca-Cola (1.5L)", 14000, "dona", 1, 1, False),
    ("Fanta (1.5L)", 14000, "dona", 1, 1, False),
    ("Sprite (1.5L)", 14000, "dona", 1, 1, False),
    ("Coca-Cola (0.5L)", 7000, "dona", 1, 1, False),
    ("Fanta (0.5L)", 7000, "dona", 1, 1, False),
    ("Choy (Qora)", 5000, "pors", 1, 1, False),
    ("Choy (Ko'k)", 5000, "pors", 1, 1, False),
    ("Limon choy", 15000, "pors", 1, 1, False),
    ("Kofe (Amerikano)", 12000, "dona", 1, 1, False),
    ("Kofe (Latte)", 18000, "dona", 1, 1, False),
    ("Kofe (Kapuchino)", 16000, "dona", 1, 1, False),
    ("Sok (Olma 1L)", 15000, "dona", 1, 1, False),
    ("Sok (Gilos 1L)", 15000, "dona", 1, 1, False),
    ("Sok (O'rik 1L)", 15000, "dona", 1, 1, False),
    ("Suv (Gazlangan 1L)", 4000, "dona", 1, 1, False),
    ("Suv (Gazsiz 1L)", 4000, "dona", 1, 1, False),
    ("Moxito", 20000, "dona", 1, 1, False),
    ("Ayron", 6000, "dona", 1, 1, False),
    ("Qimiz", 15000, "dona", 1, 1, False),
    ("Qulupnayli sheyk", 25000, "dona", 1, 1, False),
]

salatlar = [
    ("Achchiq-chuchuk", 12000, "pors", 1, 1, True),
    ("Svejiy salat", 15000, "pors", 1, 1, True),
    ("Sezar", 35000, "pors", 1, 1, True),
    ("Mujskoy kapriz", 32000, "pors", 1, 1, True),
    ("Olivye", 25000, "pors", 1, 1, True),
    ("Grek salati", 28000, "pors", 1, 1, True),
    ("Karam salati", 10000, "pors", 1, 1, True),
    ("Sabzi salati (Koreyscha)", 15000, "pors", 1, 1, True),
    ("Qo'ziqorinli salat", 30000, "pors", 1, 1, True),
    ("Tovuqli salat", 28000, "pors", 1, 1, True),
    ("Vinigret", 20000, "pors", 1, 1, True),
    ("Baxor salati", 18000, "pors", 1, 1, True),
    ("Krabli salat", 22000, "pors", 1, 1, True),
    ("Selyodka pod shuboy", 25000, "pors", 1, 1, True),
    ("Qaldirg'och uyasi", 30000, "pors", 1, 1, True),
    ("Gullar salati", 24000, "pors", 1, 1, True),
    ("Dungan salati", 20000, "pors", 1, 1, True),
    ("Pishloqli salat", 26000, "pors", 1, 1, True),
    ("Makkajo'xori salat", 18000, "pors", 1, 1, True),
    ("Go'shtli assorti", 150000, "kg", 0.5, 0.1, True),
]

shirinliklar = [
    ("Asalli tort", 15000, "pors", 1, 1, True),
    ("Napaleon", 18000, "pors", 1, 1, True),
    ("Snikers tort", 20000, "pors", 1, 1, True),
    ("Muzqaymoq (Plombir)", 10000, "pors", 1, 1, True),
    ("Muzqaymoq (Shokoladli)", 12000, "pors", 1, 1, True),
    ("Muzqaymoq (Meva assorti)", 15000, "pors", 1, 1, True),
    ("Chizkeyk (Klassik)", 22000, "dona", 1, 1, True),
    ("Chizkeyk (Qulupnayli)", 25000, "dona", 1, 1, True),
    ("Pechenye assorti", 40000, "kg", 0.5, 0.1, True),
    ("Shokoladli pechenye", 45000, "kg", 0.5, 0.1, True),
    ("Shakolatlar", 120000, "kg", 0.1, 0.1, False),
    ("Marmelad", 80000, "kg", 0.2, 0.1, True),
    ("Eksler (Krem) - 100gr", 8000, "gr", 200, 100, True),
    ("Makkaron (Pechenye)", 15000, "dona", 2, 1, True),
    ("Keks", 12000, "dona", 1, 1, True),
    ("Rulet (Meva)", 18000, "pors", 1, 1, True),
    ("Paxlava", 25000, "pors", 1, 1, True),
    ("Chak-chak", 60000, "kg", 0.3, 0.1, True),
    ("Yong'oqli pishiriq", 85000, "kg", 0.2, 0.1, True),
    ("Xolvaytar", 15000, "pors", 1, 1, True),
]

# print("INSERT INTO categories (id, name, image_url, is_user_controlled) VALUES ")
# cat_values = []
# for c in categories:
#     cat_values.append(f"({c[0]}, '{c[1]}', '{c[2]}', true)")
# print(",\n".join(cat_values) + ";\n")

print("INSERT INTO products (category_id, name, description, price, image_url, is_active, unit, min_quantity, quantity_step, has_mandatory_container) VALUES ")
prod_values = []
all_prods = [
    (10, milliy, "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400"),
    (11, fast_food, "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400"),
    (12, ichimliklar, "https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400"),
    (13, salatlar, "https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400"),
    (14, shirinliklar, "https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400")
]

for cat_id, prod_list, img in all_prods:
    for p in prod_list:
        p_name = p[0].replace("'", "''")
        prod_values.append(f"({cat_id}, '{p_name}', 'Mijoz uchun namunaviy mahsulot', {p[1]}, '{img}', true, '{p[2]}', {p[3]}, {p[4]}, {str(p[5]).lower()})")

print(",\n".join(prod_values) + ";")
