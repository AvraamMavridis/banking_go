
export default class InsufficientFunds extends Error {
  readonly statusCode = 422;

  constructor(message = "Insufficient Funds") {
    super(message);
    this.name = "InsufficientFunds";

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, InsufficientFunds);
    }
  }
}
