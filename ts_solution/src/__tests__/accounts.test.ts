import Hapi from "@hapi/hapi";
import { randomUUID } from "crypto";
import { AppDataSource } from "../db/data-source";
import { Account } from "../entities/Account";
import { IdempotencyRecord } from "../entities/IdempotencyRecord";
import { createServer } from "../server";

let server: Hapi.Server;

const validAccount = {
  name: "Alice",
  surname: "Smith",
  email: "alice@example.com",
  phone: "+441234567890",
  addressLine1: "123 High Street",
  addressLine2: "Flat 4",
  city: "London",
  postcode: "SW1A 1AA",
  country: "UK",
  balance: 100,
};

beforeAll(async () => {
  server = await createServer();
  await AppDataSource.synchronize(true);
  await server.initialize();
});

afterAll(async () => {
  await server.stop();
  await AppDataSource.destroy();
});

beforeEach(async () => {
  await AppDataSource.getRepository(IdempotencyRecord).clear();
  await AppDataSource.getRepository(Account).clear();
});

describe("POST /accounts", () => {
  it("creates a new account", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: validAccount,
    });

    expect(res.statusCode).toBe(201);
    const body = JSON.parse(res.payload);
    expect(body.name).toBe("Alice");
    expect(body.surname).toBe("Smith");
    expect(body.email).toBe("alice@example.com");
    expect(body.balance).toBe(100);
    expect(body.id).toBeDefined();
  });

  it("rejects missing name", async () => {
    const { name, ...noName } = validAccount;
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: noName,
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects missing surname", async () => {
    const { surname, ...noSurname } = validAccount;
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: noSurname,
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects missing email", async () => {
    const { email, ...noEmail } = validAccount;
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: noEmail,
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects invalid email", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: { ...validAccount, email: "not-an-email" },
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects negative balance", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: { ...validAccount, email: "neg@test.com", balance: -10 },
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects missing address fields", async () => {
    const { addressLine1, city, postcode, country, ...noAddress } = validAccount;
    const res = await server.inject({
      method: "POST",
      url: "/accounts",
      headers: { "idempotency-key": randomUUID() },
      payload: { ...noAddress, email: "addr@test.com" },
    });

    expect(res.statusCode).toBe(400);
  });
});

describe("GET /accounts", () => {
  it("returns empty list when no accounts exist", async () => {
    const res = await server.inject({ method: "GET", url: "/accounts" });

    expect(res.statusCode).toBe(200);
    expect(JSON.parse(res.payload)).toEqual([]);
  });

  it("returns all accounts", async () => {
    const repo = AppDataSource.getRepository(Account);
    await repo.save([
      repo.create({ ...validAccount, email: "a@test.com" }),
      repo.create({ ...validAccount, email: "b@test.com" }),
    ]);

    const res = await server.inject({ method: "GET", url: "/accounts" });

    expect(res.statusCode).toBe(200);
    const body = JSON.parse(res.payload);
    expect(body).toHaveLength(2);
  });
});

describe("GET /accounts/{id}", () => {
  it("returns an account by id", async () => {
    const repo = AppDataSource.getRepository(Account);
    const saved = await repo.save(repo.create(validAccount));

    const res = await server.inject({ method: "GET", url: `/accounts/${saved.id}` });

    expect(res.statusCode).toBe(200);
    const body = JSON.parse(res.payload);
    expect(body.name).toBe("Alice");
    expect(body.surname).toBe("Smith");
    expect(body.balance).toBe(100);
  });

  it("returns 404 for non-existent account", async () => {
    const res = await server.inject({ method: "GET", url: "/accounts/999" });

    expect(res.statusCode).toBe(404);
  });
});

describe("POST /accounts/{id}/deposit", () => {
  it("deposits amount and stores as cents", async () => {
    const repo = AppDataSource.getRepository(Account);
    const saved = await repo.save(repo.create({ ...validAccount, balance: 0 }));

    const res = await server.inject({
      method: "POST",
      url: `/accounts/${saved.id}/deposit`,
      headers: { "idempotency-key": randomUUID() },
      payload: { amount: 2550 },
    });

    expect(res.statusCode).toBe(200);
    const body = JSON.parse(res.payload);
    expect(body.balance).toBe(2550);
  });

  it("adds to existing balance", async () => {
    const repo = AppDataSource.getRepository(Account);
    const saved = await repo.save(
      repo.create({ ...validAccount, email: "dep@test.com", balance: 1000 })
    );

    const res = await server.inject({
      method: "POST",
      url: `/accounts/${saved.id}/deposit`,
      headers: { "idempotency-key": randomUUID() },
      payload: { amount: 1000 },
    });

    expect(res.statusCode).toBe(200);
    const body = JSON.parse(res.payload);
    expect(body.balance).toBe(2000);
  });

  it("rejects negative amount", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts/1/deposit",
      headers: { "idempotency-key": randomUUID() },
      payload: { amount: -500 },
    });

    expect(res.statusCode).toBe(400);
  });

  it("rejects zero amount", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts/1/deposit",
      headers: { "idempotency-key": randomUUID() },
      payload: { amount: 0 },
    });

    expect(res.statusCode).toBe(400);
  });

  it("returns 404 for non-existent account", async () => {
    const res = await server.inject({
      method: "POST",
      url: "/accounts/999/deposit",
      headers: { "idempotency-key": randomUUID() },
      payload: { amount: 1000 },
    });

    expect(res.statusCode).toBe(404);
  });
});
