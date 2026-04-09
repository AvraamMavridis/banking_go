import "reflect-metadata";
import { DataSource } from "typeorm";
import { Account } from "../entities/Account";
import { IdempotencyRecord } from "../entities/IdempotencyRecord";

const isTest = process.env.NODE_ENV === "test";

export const AppDataSource = new DataSource({
  type: "better-sqlite3",
  database: isTest ? ":memory:" : "./dev.sqlite3",
  synchronize: false,
  entities: [Account, IdempotencyRecord],
  migrations: [__dirname + "/migrations/*.{ts,js}"],
});
