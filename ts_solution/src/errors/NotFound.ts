
export default class NotFound extends Error {
  readonly statusCode = 404;

  constructor(message = "Not Found") {
    super(message);
    this.name = "NotFound";

    // Maintains proper stack trace (V8)
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, NotFound);
    }
  }
}