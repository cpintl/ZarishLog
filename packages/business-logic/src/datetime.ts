import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";

dayjs.extend(utc);
dayjs.extend(timezone);

// PRD §7 / ARCHITECTURE §9: default display timezone and formats.
// Store all timestamps in UTC in the database; convert at the edge.
export const DEFAULT_TIMEZONE = "Asia/Dhaka"; // GMT+6
export const DEFAULT_DATE_FORMAT = "DD MMMM YYYY"; // e.g. 15 July 2026
export const DEFAULT_TIME_FORMAT = "hh:mm A"; // e.g. 02:30 PM

export function formatDate(date: Date | string, tz: string = DEFAULT_TIMEZONE): string {
  return dayjs(date).tz(tz).format(DEFAULT_DATE_FORMAT);
}

export function formatTime(date: Date | string, tz: string = DEFAULT_TIMEZONE): string {
  return dayjs(date).tz(tz).format(DEFAULT_TIME_FORMAT);
}

export function formatDateTime(date: Date | string, tz: string = DEFAULT_TIMEZONE): string {
  return `${formatDate(date, tz)}, ${formatTime(date, tz)}`;
}

export function daysUntilExpiry(expiryDate: Date | string, tz: string = DEFAULT_TIMEZONE): number {
  const now = dayjs().tz(tz).startOf("day");
  const expiry = dayjs(expiryDate).tz(tz).startOf("day");
  return expiry.diff(now, "day");
}
