import Hapi from "@hapi/hapi";
import { AppDataSource } from "./db/data-source";
import { registerAccountRoutes } from "./routes/accounts";

export async function createServer(): Promise<Hapi.Server> {
  if (!AppDataSource.isInitialized) {
    await AppDataSource.initialize();
  }

  const server = Hapi.server({
    port: process.env.PORT || 3000,
    host: "localhost",
  });

  registerAccountRoutes(server);

  return server;
}
