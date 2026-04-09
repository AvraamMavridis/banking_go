
export default class DuplicateRequest extends Error {
  readonly statusCode: number;
  readonly cachedResponse: any;

  constructor(statusCode: number, cachedResponse: any) {
    super("Duplicate request");
    this.name = "DuplicateRequest";
    this.statusCode = statusCode;
    this.cachedResponse = cachedResponse;

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, DuplicateRequest);
    }
  }
}
