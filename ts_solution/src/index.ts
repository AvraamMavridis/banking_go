import { createServer } from "./server";

async function start(): Promise<void> {
  const server = await createServer();
  await server.start();

  console.log(`Server running on ${server.info.uri}`);
}

start().catch((err) => {
  console.error(err);
  process.exit(1);
});
