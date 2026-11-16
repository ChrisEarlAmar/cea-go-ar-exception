package report

import (
	"context"
	"database/sql"
	"time"
)

const chargedSalesQuery = `
SELECT
    ROWNUM AS ID,
    OC.COMPANY_NAME AS COMPANY,
    OS.STORE_NAME AS BRANCH,
    OTH.CUSTOMER_NAME,
    OTH.CUSTOMER_ID,
    (CASE
        WHEN TO_CHAR(OTH.TRX_TYPE) = '1' THEN 'REGULAR SALES'
        WHEN TO_CHAR(OTH.TRX_TYPE) = '2' THEN 'SALES RETURN'
        WHEN TO_CHAR(OTH.TRX_TYPE) = '3' THEN 'CREDIT CUSTOMER SALE'
        WHEN TO_CHAR(OTH.TRX_TYPE) = '4' THEN 'RESERVATION'
        ELSE TO_CHAR(OTH.TRX_TYPE)
    END) AS TRANSACTION_TYPE,
    TO_CHAR(TRUNC(OTH.TRX_DATE), 'DD-MON-YY') AS TRANSACTION_DATE,
    OTH.SALES_INVOICE_RECEIPT_NO AS SALES_INVOICE_NO,
    OTH.UNIVERSAL_TRX_ID AS TRANSACTION_NUMBER,
    OTH.PO_NUM AS PO_NUMBER,
    OTH.TRX_AMOUNT AS POS_AMOUNT,
    (CASE
        WHEN TO_CHAR(OTH.TRX_STATUS) = '1' THEN 'VOIDED'
        WHEN TO_CHAR(OTH.TRX_STATUS) = '2' THEN 'VOIDED COMPLETED'
        WHEN TO_CHAR(OTH.TRX_STATUS) = '3' THEN 'COMPLETED'
        WHEN TO_CHAR(OTH.TRX_STATUS) = '4' THEN 'SUSPENDED'
        WHEN TO_CHAR(OTH.TRX_STATUS) = '5' THEN 'RESUMED'
        WHEN TO_CHAR(OTH.TRX_STATUS) = '6' THEN 'OPEN'
        ELSE TO_CHAR(OTH.TRX_STATUS)
    END) AS TRANSACTION_STATUS,
    RCT.ATTRIBUTE1
FROM OSIPOS.OSIPOS_TRX_HEADER OTH
INNER JOIN OSIPOS.OSIPOS_COMPANY OC
    ON OTH.COMPANY_ID = OC.COMPANY_ID
INNER JOIN OSIPOS.OSIPOS_STORE OS
    ON OTH.STORE_ID = OS.STORE_ID
LEFT JOIN RA_CUSTOMER_TRX_ALL RCT
    ON TO_CHAR(OTH.UNIVERSAL_TRX_ID) = TO_CHAR(RCT.ATTRIBUTE1)
WHERE
    RCT.ATTRIBUTE1 IS NULL
    AND TO_CHAR(OTH.TRX_TYPE) IN ('2', '3')
    AND TO_CHAR(OTH.TRX_STATUS) IN ('1', '2', '3')
    AND OTH.CUSTOMER_ID <> '-1'
    AND (OC.COMPANY_ID = :company OR :company IS NULL)
    AND (OS.STORE_CODE = :branch OR :branch IS NULL)
    AND TRUNC(OTH.TRX_DATE) BETWEEN TO_DATE(:dateFrom, 'DD-MON-YY') AND TO_DATE(:dateTo, 'DD-MON-YY')
ORDER BY OC.COMPANY_NAME ASC, OS.STORE_NAME ASC, TRUNC(OTH.TRX_DATE) ASC`

type ChargedSalesFilter struct {
	DateFrom  time.Time
	DateTo    time.Time
	CompanyID *int64
	StoreCode *string
}

type Result struct {
	Columns     []string         `json:"columns"`
	Rows        []map[string]any `json:"rows"`
	Count       int              `json:"count"`
	GeneratedAt time.Time        `json:"generatedAt"`
}

type Runner interface {
	ChargedSales(context.Context, ChargedSalesFilter) (Result, error)
}

type OracleRunner struct {
	db *sql.DB
}

func NewOracleRunner(db *sql.DB) *OracleRunner {
	return &OracleRunner{db: db}
}

func (r *OracleRunner) ChargedSales(ctx context.Context, filter ChargedSalesFilter) (Result, error) {
	var company any
	if filter.CompanyID != nil {
		company = *filter.CompanyID
	}

	var branch any
	if filter.StoreCode != nil {
		branch = *filter.StoreCode
	}

	rows, err := r.db.QueryContext(
		ctx,
		chargedSalesQuery,
		sql.Named("company", company),
		sql.Named("branch", branch),
		sql.Named("dateFrom", filter.DateFrom.Format("02-Jan-06")),
		sql.Named("dateTo", filter.DateTo.Format("02-Jan-06")),
	)
	if err != nil {
		return Result{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Columns:     columns,
		Rows:        make([]map[string]any, 0),
		GeneratedAt: time.Now().UTC(),
	}

	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}

		if err := rows.Scan(pointers...); err != nil {
			return Result{}, err
		}

		row := make(map[string]any, len(columns))
		for index, column := range columns {
			row[column] = jsonValue(values[index])
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return Result{}, err
	}

	result.Count = len(result.Rows)
	return result, nil
}

func jsonValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format(time.RFC3339)
	default:
		return typed
	}
}
