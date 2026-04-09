import { MigrationInterface, QueryRunner, Table } from "typeorm";

export class CreateAccounts1712563200000 implements MigrationInterface {
  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.createTable(
      new Table({
        name: "accounts",
        columns: [
          {
            name: "id",
            type: "integer",
            isPrimary: true,
            isGenerated: true,
            generationStrategy: "increment",
          },
          { name: "name", type: "varchar", length: "255" },
          { name: "surname", type: "varchar", length: "255" },
          { name: "email", type: "varchar", length: "255", isUnique: true },
          { name: "phone", type: "varchar", length: "50", isNullable: true },
          { name: "address_line1", type: "varchar", length: "255" },
          { name: "address_line2", type: "varchar", length: "255", isNullable: true },
          { name: "city", type: "varchar", length: "100" },
          { name: "postcode", type: "varchar", length: "20" },
          { name: "country", type: "varchar", length: "100" },
          { name: "balance", type: "decimal", precision: 14, scale: 2, default: 0 },
          { name: "created_at", type: "datetime", default: "CURRENT_TIMESTAMP" },
          { name: "updated_at", type: "datetime", default: "CURRENT_TIMESTAMP" },
        ],
      }),
      true,
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.dropTable("accounts");
  }
}
