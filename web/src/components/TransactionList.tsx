import { FC } from 'react';
import { Transaction } from '../types';

interface TransactionListProps {
  transactions: Transaction[];
}

export const TransactionList: FC<TransactionListProps> = ({ transactions }) => {
  return (
    <div>
      <h2>Transactions</h2>
      <ul>
        {transactions.map((transaction) => (
          <li key={transaction.id}>
            {transaction.description} - {transaction.amount} {transaction.currency}
          </li>
        ))}
      </ul>
    </div>
  );
}; 