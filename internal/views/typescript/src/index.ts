import { phone } from "phone";

globalThis.toE164 = function (phoneNumber: string, country = "ES") {
  return (
    phone(phoneNumber, { country, validateMobilePrefix: false }).phoneNumber ||
    phoneNumber
  );
};
