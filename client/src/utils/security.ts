/**
 * Security utilities for input validation and sanitization
 */

// XSS prevention - sanitize HTML content
export function sanitizeHtml(input: string): string {
  const div = document.createElement('div');
  div.textContent = input;
  return div.innerHTML;
}

// Input validation patterns
export const VALIDATION_PATTERNS = {
  username: /^[a-zA-Z0-9_-]{3,20}$/,
  email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  password: /^.{8,}$/, // Minimum 8 characters
  phone: /^\+?[1-9]\d{1,14}$/,
  url: /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$/
} as const;

// Validate input against common injection patterns
export function validateInput(input: string, type: keyof typeof VALIDATION_PATTERNS): boolean {
  if (!input || typeof input !== 'string') return false;
  return VALIDATION_PATTERNS[type].test(input.trim());
}

// Rate limiting helper (client-side)
class RateLimiter {
  private attempts = new Map<string, number[]>();
  
  isAllowed(key: string, maxAttempts: number = 5, windowMs: number = 60000): boolean {
    const now = Date.now();
    const windowStart = now - windowMs;
    
    const keyAttempts = this.attempts.get(key) || [];
    const validAttempts = keyAttempts.filter(timestamp => timestamp > windowStart);
    
    if (validAttempts.length >= maxAttempts) {
      return false;
    }
    
    validAttempts.push(now);
    this.attempts.set(key, validAttempts);
    return true;
  }
  
  reset(key: string): void {
    this.attempts.delete(key);
  }
}

export const rateLimiter = new RateLimiter();

// CSRF token helper (for future implementation)
export function generateCSRFToken(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
}