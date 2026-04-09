import {
  Entity,
  PrimaryColumn,
  Column,
  CreateDateColumn,
} from "typeorm";

@Entity("idempotency_records")
export class IdempotencyRecord {
  @PrimaryColumn({ length: 255 })
  key!: string;

  @Column({ type: "text" })
  response!: string;

  @Column({ type: "integer" })
  statusCode!: number;

  @CreateDateColumn({ name: "created_at" })
  createdAt!: Date;
}
