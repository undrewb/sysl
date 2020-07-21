/*TITLE : */
/* ---------------------------------------------
Autogenerated script from sysl
--------------------------------------------- */


/*-----------------------Relation Model : BankModel-----------------------------------------------*/
CREATE TABLE Branch(
  branch_id integer,
  branch_name varchar (50),
  branch_address varchar (50),
  CONSTRAINT BRANCH_PK PRIMARY KEY(branch_id)
);
CREATE TABLE Account(
  account_number integer,
  account_type varchar (50),
  account_status varchar (50),
  account_balance integer,
  CONSTRAINT ACCOUNT_PK PRIMARY KEY(account_number)
);
CREATE TABLE Customer(
  customer_id integer,
  customer_name varchar (50),
  customer_address varchar (50),
  customer_dob date,
  branch_id integer,
  CONSTRAINT CUSTOMER_PK PRIMARY KEY(customer_id),
  CONSTRAINT CUSTOMER_BRANCH_ID_FK FOREIGN KEY(branch_id) REFERENCES Branch (branch_id)
);
CREATE TABLE Transaction(
  transaction_id integer,
  transaction_type varchar (50),
  transaction_date_time date,
  transaction_amount integer,
  from_account_number integer,
  to_account_number integer,
  CONSTRAINT TRANSACTION_PK PRIMARY KEY(transaction_id),
  CONSTRAINT TRANSACTION_FROM_ACCOUNT_NUMBER_FK FOREIGN KEY(from_account_number) REFERENCES Account (account_number),
  CONSTRAINT TRANSACTION_TO_ACCOUNT_NUMBER_FK FOREIGN KEY(to_account_number) REFERENCES Account (account_number)
);
CREATE TABLE CustomerAccount(
  customer_id integer,
  account_number integer,

  CONSTRAINT CUSTOMERACCOUNT_CUSTOMER_ID_FK FOREIGN KEY(customer_id) REFERENCES Customer (customer_id),
  CONSTRAINT CUSTOMERACCOUNT_ACCOUNT_NUMBER_FK FOREIGN KEY(account_number) REFERENCES Account (account_number)
);