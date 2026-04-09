import { MigrationInterface, QueryRunner, Table } from "typeorm";

export class CreateIdempotencyRecords1712563200001 implements MigrationInterface {
  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.createTable(
      new Table({
        name: "idempotency_records",
        columns: [
          {
            name: "key",
            type: "varchar",
            length: "255",
            isPrimary: true,
          },
          { name: "response", type: "text" },
          { name: "status_code", type: "integer" },
          { name: "created_at", type: "datetime", default: "CURRENT_TIMESTAMP" },
        ],
      }),
      true,
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.dropTable("idempotency_records");
  }
}
