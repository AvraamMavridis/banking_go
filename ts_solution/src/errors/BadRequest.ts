
export default class BadRequest extends Error {
  readonly statusCode = 400;

  constructor(message = "Bad Request") {
    super(message);
    this.name = "BadRequest";

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, BadRequest);
    }
  }
}
