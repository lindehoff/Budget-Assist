export interface Transaction {
  id: number;
  amount: number;
  currency: string;
  description: string;
  transactionDate: Date;
  categoryId?: number;
  subcategoryId?: number;
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  isActive: boolean;
  typeId: number;
}

export interface CategoryType {
  id: number;
  name: string;
  description?: string;
  isMultiple: boolean;
} 