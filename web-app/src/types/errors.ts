export class NetworkError extends Error {
  cause?: unknown;
  status: number;

  constructor(message: string, status: number, cause?: unknown) {
    super(message);
    this.status = status;
    this.cause = cause;
  }
}
