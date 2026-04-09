import { AppDataSource } from "./db/data-source";

AppDataSource.initialize()
  .then(async () => {
    await AppDataSource.runMigrations();
    console.log("Migrations completed successfully");
    await AppDataSource.destroy();
    process.exit(0);
  })
  .catch((err) => {
    console.error("Migration failed:", err);
    process.exit(1);
  });
