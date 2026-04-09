import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from "typeorm";

@Entity("accounts")
export class Account {
  @PrimaryGeneratedColumn()
  id!: number;

  @Column({ length: 255 })
  name!: string;

  @Column({ length: 255 })
  surname!: string;

  @Column({ length: 255, unique: true })
  email!: string;

  @Column({ type: "varchar", length: 50, nullable: true })
  phone!: string | null;

  @Column({ name: "address_line1", length: 255 })
  addressLine1!: string;

  @Column({ name: "address_line2", type: "varchar", length: 255, nullable: true })
  addressLine2!: string | null;

  @Column({ length: 100 })
  city!: string;

  @Column({ length: 20 })
  postcode!: string;

  @Column({ length: 100 })
  country!: string;

  @Column({ type: "integer", default: 0 })
  balance!: number;

  @Column({ length: 3, default: "EUR" })
  currency!: string;

  @CreateDateColumn({ name: "created_at" })
  createdAt!: Date;

  @UpdateDateColumn({ name: "updated_at" })
  updatedAt!: Date;
}
