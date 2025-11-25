import { phone } from "phone";

globalThis.toE164 = function (phoneNumber: string, country = "ES") {
  return (
    phone(phoneNumber, { country, validateMobilePrefix: false }).phoneNumber ||
    phoneNumber
  );
};

// Reference: https://getbutterfly.com/generate-a-password-using-vanilla-javascript/
globalThis.generateRandomPassword = function (length: number = 12) {
  // Define character sets
  const lowercaseChars = 'abcdefghijklmnopqrstuvwxyz';
  const uppercaseChars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  const numberChars = '0123456789';
  const symbolChars = '!@#$%^&*()-_=+[]{}|;:,.<>?';

  // Combine all character sets
  const allChars = lowercaseChars + uppercaseChars + numberChars + symbolChars;

  let password = '';

  // Ensure the password includes at least one character from each set
  password += lowercaseChars.charAt(Math.floor(Math.random() * lowercaseChars.length));
  password += uppercaseChars.charAt(Math.floor(Math.random() * uppercaseChars.length));
  password += numberChars.charAt(Math.floor(Math.random() * numberChars.length));
  password += symbolChars.charAt(Math.floor(Math.random() * symbolChars.length));

  // Fill the remaining characters
  for (let i = password.length; i < length; i++) {
    password += allChars.charAt(Math.floor(Math.random() * allChars.length));
  }

  return password
}