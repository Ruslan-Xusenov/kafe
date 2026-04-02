export const validatePhone = (phone) => {
  const re = /^\+998[0-9]{9}$/;
  if (!phone) return "Telefon raqami kiritilishi shart";
  if (!re.test(phone.replace(/\s/g, ''))) return "Telefon raqami noto'g'ri (+998XXXXXXXXX ko'rinishida bo'lishi kerak)";
  return null;
};

export const validateFullName = (name) => {
  if (!name) return "Ism sharif kiritilishi shart";
  if (name.trim().length < 3) return "Ism sharif kamida 3ta belgi bo'lishi kerak";
  return null;
};

export const validatePassword = (password) => {
  if (!password) return "Parol kiritilishi shart";
  if (password.length < 6) return "Parol kamida 6ta belgi bo'lishi kerak";
  return null;
};

export const validatePrice = (price) => {
  const p = parseFloat(price);
  if (isNaN(p)) return "Narxi son bo'lishi shart";
  if (p <= 0) return "Narxi 0 dan katta bo'lishi shart";
  return null;
};

export const validateAddress = (address) => {
  if (!address) return "Manzil kiritilishi shart";
  if (address.trim().length < 5) return "Manzilni batafsilroq kiriting";
  return null;
};

export const validateNotEmpty = (val, fieldName = "Maydon") => {
  if (!val || val.toString().trim() === "") return `${fieldName} to'ldirilishi shart`;
  return null;
};
