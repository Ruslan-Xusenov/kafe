import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';

// ─── Translations ───────────────────────────────────────────────────────────
export const LANG = {
  uz: {
    // Login
    login: "Kirish",
    phone: "Telefon raqam",
    password: "Parol",
    loginBtn: "Tizimga kirish",
    loginError: "Kirish xatosi",

    // Dashboard
    waiterPanel: "Ofitsant paneli",
    tables: "Stollar",
    free: "Bo'sh",
    occupied: "Band",
    table: "Stol",
    refresh: "Yangilash",
    logout: "Chiqish",

    // Table detail
    tableDetail: "Stol",
    activeOrder: "Faol buyurtma",
    noActiveOrder: "Faol buyurtma yo'q",
    addItems: "Mahsulot qo'shish",
    closeTable: "Stol yopish",
    transferTable: "Ko'chirish",
    serviceFee: "Xizmat haqi",
    cancel: "Bekor qilish",
    confirm: "Tasdiqlash",
    sum: "so'm",
    qty: "dona",
    total: "Jami",
    subtotal: "Oraliq",
    paymentMethod: "To'lov usuli",
    cash: "Naqd",
    card: "Karta",
    click: "Click/Payme",
    nasiya: "Nasiya",
    mixed: "Aralash",
    close: "Yopish",
    back: "Orqaga",
    newOrder: "Yangi buyurtma",
    orderSent: "Buyurtma yuborildi!",
    orderSentMsg: "Buyurtma muvaffaqiyatli oshxonaga yuborildi.",
    errorTitle: "Xatolik",
    enterAmount: "Summa kiriting",
    enterValidPercent: "0 dan 100 gacha son kiriting",

    // New order
    menu: "Menyu",
    search: "Qidirish...",
    allCategories: "Barchasi",
    cart: "Savat",
    cartEmpty: "Savat bo'sh",
    sendOrder: "Buyurtma yuborish",
    orderTotal: "Jami",

    // History
    history: "Tarix",
    ordersHistory: "Buyurtmalar tarixi",
    noHistory: "Buyurtmalar tarixi yo'q",
    status_new: "Yangi",
    status_preparing: "Tayyorlanmoqda",
    status_ready: "Tayyor",
    status_on_way: "Yo'lda",
    status_delivered: "Yetkazildi",
    status_cancelled: "Bekor qilindi",

    // Profile
    profile: "Profil",
    language: "Til",
    chooseLang: "Tilni tanlang",
    version: "Versiya",

    // Transfer
    transferTitle: "Stol ko'chirish",
    selectTargetTable: "Manzil stolni tanlang",
    transferConfirm: "Ko'chirmoqchimisiz?",

    // Service fee
    serviceFeeTitle: "Xizmat haqi (%)",
    serviceFeeSet: "Belgilash",

    // Cancel
    cancelItem: "Mahsulotni bekor qilish",
    cancelQty: "Bekor qilish miqdori",
  },
  ru: {
    // Login
    login: "Вход",
    phone: "Номер телефона",
    password: "Пароль",
    loginBtn: "Войти в систему",
    loginError: "Ошибка входа",

    // Dashboard
    waiterPanel: "Панель официанта",
    tables: "Столы",
    free: "Свободен",
    occupied: "Занят",
    table: "Стол",
    refresh: "Обновить",
    logout: "Выйти",

    // Table detail
    tableDetail: "Стол",
    activeOrder: "Активный заказ",
    noActiveOrder: "Нет активного заказа",
    addItems: "Добавить блюда",
    closeTable: "Закрыть стол",
    transferTable: "Перенести",
    serviceFee: "Сервисный сбор",
    cancel: "Отменить",
    confirm: "Подтвердить",
    sum: "сум",
    qty: "шт",
    total: "Итого",
    subtotal: "Промежуток",
    paymentMethod: "Метод оплаты",
    cash: "Наличные",
    card: "Карта",
    click: "Click/Payme",
    nasiya: "В долг",
    mixed: "Смешанный",
    close: "Закрыть",
    back: "Назад",
    newOrder: "Новый заказ",
    orderSent: "Заказ отправлен!",
    orderSentMsg: "Заказ успешно отправлен на кухню.",
    errorTitle: "Ошибка",
    enterAmount: "Введите сумму",
    enterValidPercent: "Введите число от 0 до 100",

    // New order
    menu: "Меню",
    search: "Поиск...",
    allCategories: "Все",
    cart: "Корзина",
    cartEmpty: "Корзина пуста",
    sendOrder: "Отправить заказ",
    orderTotal: "Итого",

    // History
    history: "История",
    ordersHistory: "История заказов",
    noHistory: "История заказов пуста",
    status_new: "Новый",
    status_preparing: "Готовится",
    status_ready: "Готов",
    status_on_way: "В пути",
    status_delivered: "Доставлен",
    status_cancelled: "Отменён",

    // Profile
    profile: "Профиль",
    language: "Язык",
    chooseLang: "Выберите язык",
    version: "Версия",

    // Transfer
    transferTitle: "Перенос стола",
    selectTargetTable: "Выберите стол назначения",
    transferConfirm: "Вы хотите перенести заказ?",

    // Service fee
    serviceFeeTitle: "Сервисный сбор (%)",
    serviceFeeSet: "Установить",

    // Cancel
    cancelItem: "Отменить блюдо",
    cancelQty: "Кол-во для отмены",
  },
};

export const useLangStore = create((set, get) => ({
  lang: 'uz',
  t: LANG['uz'],

  setLang: async (lang) => {
    await AsyncStorage.setItem('app_lang', lang);
    set({ lang, t: LANG[lang] });
  },

  loadLang: async () => {
    const saved = await AsyncStorage.getItem('app_lang');
    const lang = saved || 'uz';
    set({ lang, t: LANG[lang] });
  },
}));
