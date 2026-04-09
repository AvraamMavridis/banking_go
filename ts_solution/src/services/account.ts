import { AppDataSource } from "../db/data-source";
import { Account } from "../entities/Account";
import { IdempotencyRecord } from "../entities/IdempotencyRecord";
import NotFound from "../errors/NotFound";
import InsufficientFunds from "../errors/InsufficientFunds";
import BadRequest from "../errors/BadRequest";
import DuplicateRequest from "../errors/DuplicateRequest";

class AccountService {
    repo: any;

    constructor() {
        this.repo = AppDataSource.getRepository(Account);
    }

    async findById(id: number): Promise<Account> {
        const account = await this.repo.findOneBy({ id: Number(id) });
        if(!account) {
            throw new NotFound();
        }
        return account;
    }

    async create(idempotencyKey: string, payload: Partial<Account>): Promise<Account> {
        await this.ensureIdempotency(idempotencyKey);
        const account = this.repo.create(payload);
        const saved = await this.repo.save(account);
        await this.saveIdempotency(idempotencyKey, 201, saved);
        return saved;
    }

    async deposit(idempotencyKey: string, id: number, amount: number): Promise<Account> {
        await this.ensureIdempotency(idempotencyKey);

        const saved = await AppDataSource.transaction(async (manager) => {
            const account = await manager.findOneBy(Account, { id });
            if (!account) throw new NotFound();

            account.balance += amount;
            return manager.save(account);
        });

        await this.saveIdempotency(idempotencyKey, 200, saved);
        return saved;
    }

    async transfer(idempotencyKey: string, fromId: number, toId: number, amount: number): Promise<{ from: Account; to: Account }> {
        if (fromId === toId) {
            throw new BadRequest("Cannot transfer to the same account");
        }

        await this.ensureIdempotency(idempotencyKey);

        const result = await AppDataSource.transaction(async (manager) => {
            const from = await manager.findOneBy(Account, { id: fromId });
            if (!from) throw new NotFound("Source account not found");

            const to = await manager.findOneBy(Account, { id: toId });
            if (!to) throw new NotFound("Destination account not found");

            if (from.balance < amount) {
                throw new InsufficientFunds();
            }

            from.balance -= amount;
            to.balance += amount;

            await manager.save(from);
            await manager.save(to);

            return { from, to };
        });

        await this.saveIdempotency(idempotencyKey, 200, result);
        return result;
    }

    private async ensureIdempotency(key: string): Promise<void> {
        const record = await AppDataSource.getRepository(IdempotencyRecord).findOneBy({ key });
        if (record) {
            throw new DuplicateRequest(record.statusCode, JSON.parse(record.response));
        }
    }

    private async saveIdempotency(key: string, statusCode: number, response: any): Promise<void> {
        const record = new IdempotencyRecord();
        record.key = key;
        record.statusCode = statusCode;
        record.response = JSON.stringify(response);
        await AppDataSource.getRepository(IdempotencyRecord).save(record);
    }
}

const accountService = new AccountService();

export default accountService;
