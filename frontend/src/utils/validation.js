export const validatePhone = (phone) => {
  const re = /^\+998[0-9]{9}$/;
  if (!phone) return "Необходимо ввести номер телефона";
  if (!re.test(phone.replace(/\s/g, ''))) return "Неверный номер телефона (должен быть в формате +998XXXXXXXXX)";
  return null;
};

export const validateFullName = (name) => {
  if (!name) return "Необходимо ввести ФИО";
  if (name.trim().length < 3) return "ФИО должно содержать не менее 3 символов";
  return null;
};

export const validatePassword = (password) => {
  if (!password) return "Необходимо ввести пароль";
  if (password.length < 6) return "Пароль должен содержать не менее 6 символов";
  return null;
};

export const validatePrice = (price) => {
  const p = parseFloat(price);
  if (isNaN(p)) return "Цена должна быть числом";
  if (p <= 0) return "Цена должна быть больше 0";
  return null;
};

export const validateAddress = (address) => {
  if (!address) return "Необходимо ввести адрес";
  if (address.trim().length < 5) return "Введите более подробный адрес";
  return null;
};

export const validateNotEmpty = (val, fieldName = "Поле") => {
  if (!val || val.toString().trim() === "") return `Поле ${fieldName} обязательно для заполнения`;
  return null;
};
