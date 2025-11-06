# MongoDB Database Indexes

This document provides comprehensive indexing strategies for all collections in the LUJAY Assessment application. Proper indexing is crucial for query performance, especially as data volume grows.

## Table of Contents
- [Users Collection](#users-collection)
- [Vehicles Collection](#vehicles-collection)
- [Inspections Collection](#inspections-collection)
- [Transactions Collection](#transactions-collection)
- [General Index Guidelines](#general-index-guidelines)

---

## Users Collection

### Primary Indexes

```javascript
// Unique index on email (for authentication and user lookups)
db.users.createIndex({ email: 1 }, { unique: true, name: "idx_users_email_unique" })

// Index on role (for role-based queries and RBAC)
db.users.createIndex({ role: 1 }, { name: "idx_users_role" })

// Compound index on role and createdAt (for admin dashboards)
db.users.createIndex({ role: 1, createdAt: -1 }, { name: "idx_users_role_created" })
```

### Query Examples
```javascript
// Find user by email (uses idx_users_email_unique)
db.users.findOne({ email: "user@example.com" })

// Find all dealers (uses idx_users_role)
db.users.find({ role: "dealer" })

// Find recently registered admins (uses idx_users_role_created)
db.users.find({ role: "admin" }).sort({ createdAt: -1 })
```

---

## Vehicles Collection

### Primary Indexes

```javascript
// Index on ownerId (for ownership queries - get my vehicles)
db.vehicles.createIndex({ ownerId: 1 }, { name: "idx_vehicles_ownerid" })

// Index on status (for filtering active/sold/archived vehicles)
db.vehicles.createIndex({ status: 1 }, { name: "idx_vehicles_status" })

// Compound index on status and createdAt (for listing recent active vehicles)
db.vehicles.createIndex({ status: 1, createdAt: -1 }, { name: "idx_vehicles_status_created" })

// Compound index on make, model, and year (for search and filtering)
db.vehicles.createIndex({ make: 1, model: 1, year: -1 }, { name: "idx_vehicles_make_model_year" })

// Index on price (for price range queries and sorting)
db.vehicles.createIndex({ price: 1 }, { name: "idx_vehicles_price" })

// Index on year (for year range queries)
db.vehicles.createIndex({ year: 1 }, { name: "idx_vehicles_year" })

// Compound index on status and price (for active vehicles by price)
db.vehicles.createIndex({ status: 1, price: 1 }, { name: "idx_vehicles_status_price" })

// Compound index on ownerId and status (for owner's active vehicles)
db.vehicles.createIndex({ ownerId: 1, status: 1 }, { name: "idx_vehicles_owner_status" })
```

### Text Search Index

```javascript
// Text index for full-text search on make, model, and description
db.vehicles.createIndex(
  { 
    make: "text", 
    model: "text", 
    "meta.description": "text" 
  }, 
  { 
    name: "idx_vehicles_text_search",
    weights: {
      make: 10,
      model: 8,
      "meta.description": 5
    }
  }
)
```

### Query Examples
```javascript
// Get user's vehicles (uses idx_vehicles_ownerid)
db.vehicles.find({ ownerId: ObjectId("...") })

// Get active vehicles (uses idx_vehicles_status)
db.vehicles.find({ status: "active" })

// Search by make and model (uses idx_vehicles_make_model_year)
db.vehicles.find({ make: "Toyota", model: "Camry" }).sort({ year: -1 })

// Price range query (uses idx_vehicles_price)
db.vehicles.find({ price: { $gte: 20000, $lte: 35000 } })

// Active vehicles sorted by price (uses idx_vehicles_status_price)
db.vehicles.find({ status: "active" }).sort({ price: 1 })

// Text search (uses idx_vehicles_text_search)
db.vehicles.find({ $text: { $search: "Toyota Camry sedan" } })
```

---

## Inspections Collection

### Primary Indexes

```javascript
// Index on vehicleId (for getting all inspections for a vehicle)
db.inspections.createIndex({ vehicleId: 1 }, { name: "idx_inspections_vehicleid" })

// Index on inspectorId (for getting inspector's inspections)
db.inspections.createIndex({ inspectorId: 1 }, { name: "idx_inspections_inspectorid" })

// Index on status (for filtering by inspection status)
db.inspections.createIndex({ status: 1 }, { name: "idx_inspections_status" })

// Compound index on inspectorId and scheduledAt (for inspector's schedule)
db.inspections.createIndex({ inspectorId: 1, scheduledAt: -1 }, { name: "idx_inspections_inspector_scheduled" })

// Compound index on vehicleId and createdAt (for vehicle inspection history)
db.inspections.createIndex({ vehicleId: 1, createdAt: -1 }, { name: "idx_inspections_vehicle_created" })

// Compound index on status and scheduledAt (for upcoming inspections)
db.inspections.createIndex({ status: 1, scheduledAt: 1 }, { name: "idx_inspections_status_scheduled" })

// Index on scheduledAt (for date-based queries)
db.inspections.createIndex({ scheduledAt: 1 }, { name: "idx_inspections_scheduled" })

// Compound index on status and completedAt (for completed inspections timeline)
db.inspections.createIndex({ status: 1, completedAt: -1 }, { name: "idx_inspections_status_completed" })
```

### Query Examples
```javascript
// Get all inspections for a vehicle (uses idx_inspections_vehicleid)
db.inspections.find({ vehicleId: ObjectId("...") })

// Get inspector's scheduled inspections (uses idx_inspections_inspector_scheduled)
db.inspections.find({ 
  inspectorId: ObjectId("..."),
  status: "scheduled"
}).sort({ scheduledAt: 1 })

// Get upcoming scheduled inspections (uses idx_inspections_status_scheduled)
db.inspections.find({ 
  status: "scheduled",
  scheduledAt: { $gte: new Date() }
}).sort({ scheduledAt: 1 })

// Get completed inspections (uses idx_inspections_status_completed)
db.inspections.find({ status: "completed" }).sort({ completedAt: -1 })
```

---

## Transactions Collection

### Primary Indexes

```javascript
// Index on vehicleId (for getting all transactions for a vehicle)
db.transactions.createIndex({ vehicleId: 1 }, { name: "idx_transactions_vehicleid" })

// Index on sellerId (for seller's transaction history)
db.transactions.createIndex({ sellerId: 1 }, { name: "idx_transactions_sellerid" })

// Index on buyerId (for buyer's transaction history)
db.transactions.createIndex({ buyerId: 1 }, { name: "idx_transactions_buyerid" })

// Index on status (for filtering by transaction status)
db.transactions.createIndex({ status: 1 }, { name: "idx_transactions_status" })

// Compound index on sellerId and status (for seller's active transactions)
db.transactions.createIndex({ sellerId: 1, status: 1 }, { name: "idx_transactions_seller_status" })

// Compound index on buyerId and status (for buyer's active transactions)
db.transactions.createIndex({ buyerId: 1, status: 1 }, { name: "idx_transactions_buyer_status" })

// Compound index on status and createdAt (for recent transactions by status)
db.transactions.createIndex({ status: 1, createdAt: -1 }, { name: "idx_transactions_status_created" })

// Compound index on vehicleId and createdAt (for vehicle transaction history)
db.transactions.createIndex({ vehicleId: 1, createdAt: -1 }, { name: "idx_transactions_vehicle_created" })

// Index on inspectionId (for linking to inspection)
db.transactions.createIndex({ inspectionId: 1 }, { sparse: true, name: "idx_transactions_inspectionid" })

// Index on amount (for financial queries and reporting)
db.transactions.createIndex({ amount: 1 }, { name: "idx_transactions_amount" })

// Compound index on status and completedAt (for completed transactions timeline)
db.transactions.createIndex({ status: 1, completedAt: -1 }, { sparse: true, name: "idx_transactions_status_completed" })

// Index on paymentMethod (for payment analytics)
db.transactions.createIndex({ paymentMethod: 1 }, { name: "idx_transactions_payment_method" })
```

### Query Examples
```javascript
// Get all transactions for a vehicle (uses idx_transactions_vehicleid)
db.transactions.find({ vehicleId: ObjectId("...") })

// Get seller's transactions (uses idx_transactions_sellerid)
db.transactions.find({ sellerId: ObjectId("...") }).sort({ createdAt: -1 })

// Get buyer's pending transactions (uses idx_transactions_buyer_status)
db.transactions.find({ 
  buyerId: ObjectId("..."),
  status: "pending"
})

// Get completed transactions (uses idx_transactions_status_completed)
db.transactions.find({ 
  status: "completed"
}).sort({ completedAt: -1 })

// Get transactions by inspection (uses idx_transactions_inspectionid)
db.transactions.find({ inspectionId: ObjectId("...") })

// Financial report by payment method (uses idx_transactions_payment_method)
db.transactions.aggregate([
  { $group: { _id: "$paymentMethod", total: { $sum: "$amount" } } }
])
```

---

## Uploads Collection (Future)

If implementing file uploads for vehicle images or inspection reports:

```javascript
// Index on vehicleId (for getting all uploads for a vehicle)
db.uploads.createIndex({ vehicleId: 1 }, { name: "idx_uploads_vehicleid" })

// Index on inspectionId (for getting inspection report uploads)
db.uploads.createIndex({ inspectionId: 1 }, { sparse: true, name: "idx_uploads_inspectionid" })

// Index on uploadedBy (for user's upload history)
db.uploads.createIndex({ uploadedBy: 1 }, { name: "idx_uploads_uploadedby" })

// Compound index on vehicleId and type (for specific file types per vehicle)
db.uploads.createIndex({ vehicleId: 1, type: 1 }, { name: "idx_uploads_vehicle_type" })
```

---

## General Index Guidelines

### Index Creation Best Practices

1. **Create indexes before production deployment**
   - Run index creation commands during deployment scripts
   - Monitor index creation progress for large collections

2. **Index Cardinality**
   - High cardinality fields (email, ObjectId) make excellent index candidates
   - Low cardinality fields (status, role) benefit from compound indexes

3. **Compound Index Order**
   - Place equality conditions first (e.g., `status: "active"`)
   - Place sort fields last (e.g., `createdAt: -1`)
   - Place range queries in between

4. **Index Size Monitoring**
   ```javascript
   // Check index sizes
   db.vehicles.stats().indexSizes
   db.transactions.stats().indexSizes
   ```

5. **Query Performance Analysis**
   ```javascript
   // Explain a query to see which indexes are used
   db.vehicles.find({ status: "active" }).sort({ price: 1 }).explain("executionStats")
   ```

### Index Maintenance

```javascript
// Rebuild indexes (useful if data is corrupted or after major schema changes)
db.vehicles.reIndex()

// Drop unused index
db.vehicles.dropIndex("idx_name")

// List all indexes
db.vehicles.getIndexes()
```

### Monitoring Index Usage

```javascript
// Get index usage statistics (MongoDB 3.2+)
db.vehicles.aggregate([{ $indexStats: {} }])
```

### Background Index Creation

For large collections, create indexes in the background to avoid blocking:

```javascript
db.vehicles.createIndex({ price: 1 }, { background: true, name: "idx_vehicles_price" })
```

**Note:** In MongoDB 4.2+, all index builds use an optimized build process automatically.

---

## Deployment Script

Create a script to initialize all indexes at once:

```javascript
// init_indexes.js

// Users collection
db.users.createIndex({ email: 1 }, { unique: true, name: "idx_users_email_unique" });
db.users.createIndex({ role: 1 }, { name: "idx_users_role" });
db.users.createIndex({ role: 1, createdAt: -1 }, { name: "idx_users_role_created" });

// Vehicles collection
db.vehicles.createIndex({ ownerId: 1 }, { name: "idx_vehicles_ownerid" });
db.vehicles.createIndex({ status: 1 }, { name: "idx_vehicles_status" });
db.vehicles.createIndex({ status: 1, createdAt: -1 }, { name: "idx_vehicles_status_created" });
db.vehicles.createIndex({ make: 1, model: 1, year: -1 }, { name: "idx_vehicles_make_model_year" });
db.vehicles.createIndex({ price: 1 }, { name: "idx_vehicles_price" });
db.vehicles.createIndex({ year: 1 }, { name: "idx_vehicles_year" });
db.vehicles.createIndex({ status: 1, price: 1 }, { name: "idx_vehicles_status_price" });
db.vehicles.createIndex({ ownerId: 1, status: 1 }, { name: "idx_vehicles_owner_status" });
db.vehicles.createIndex({ make: "text", model: "text", "meta.description": "text" }, { name: "idx_vehicles_text_search" });

// Inspections collection
db.inspections.createIndex({ vehicleId: 1 }, { name: "idx_inspections_vehicleid" });
db.inspections.createIndex({ inspectorId: 1 }, { name: "idx_inspections_inspectorid" });
db.inspections.createIndex({ status: 1 }, { name: "idx_inspections_status" });
db.inspections.createIndex({ inspectorId: 1, scheduledAt: -1 }, { name: "idx_inspections_inspector_scheduled" });
db.inspections.createIndex({ vehicleId: 1, createdAt: -1 }, { name: "idx_inspections_vehicle_created" });
db.inspections.createIndex({ status: 1, scheduledAt: 1 }, { name: "idx_inspections_status_scheduled" });
db.inspections.createIndex({ scheduledAt: 1 }, { name: "idx_inspections_scheduled" });
db.inspections.createIndex({ status: 1, completedAt: -1 }, { name: "idx_inspections_status_completed" });

// Transactions collection
db.transactions.createIndex({ vehicleId: 1 }, { name: "idx_transactions_vehicleid" });
db.transactions.createIndex({ sellerId: 1 }, { name: "idx_transactions_sellerid" });
db.transactions.createIndex({ buyerId: 1 }, { name: "idx_transactions_buyerid" });
db.transactions.createIndex({ status: 1 }, { name: "idx_transactions_status" });
db.transactions.createIndex({ sellerId: 1, status: 1 }, { name: "idx_transactions_seller_status" });
db.transactions.createIndex({ buyerId: 1, status: 1 }, { name: "idx_transactions_buyer_status" });
db.transactions.createIndex({ status: 1, createdAt: -1 }, { name: "idx_transactions_status_created" });
db.transactions.createIndex({ vehicleId: 1, createdAt: -1 }, { name: "idx_transactions_vehicle_created" });
db.transactions.createIndex({ inspectionId: 1 }, { sparse: true, name: "idx_transactions_inspectionid" });
db.transactions.createIndex({ amount: 1 }, { name: "idx_transactions_amount" });
db.transactions.createIndex({ status: 1, completedAt: -1 }, { sparse: true, name: "idx_transactions_status_completed" });
db.transactions.createIndex({ paymentMethod: 1 }, { name: "idx_transactions_payment_method" });

print("All indexes created successfully!");
```

### Run the script:
```bash
mongosh mongodb://localhost:27017/lujay_db init_indexes.js
```

---

## Performance Recommendations

1. **Regular Monitoring**: Check slow query logs and use MongoDB Atlas performance advisor
2. **Index Coverage**: Aim for 95%+ index coverage on frequently queried fields
3. **Write Performance**: Balance read optimization with write performance (each index adds write overhead)
4. **Memory**: Ensure indexes fit in RAM for optimal performance
5. **Compound Indexes**: Prefer compound indexes over multiple single-field indexes when queries use multiple fields

---

## Index Impact Analysis

| Collection    | Indexes | Estimated Size (1M docs) | Write Overhead | Read Improvement |
|---------------|---------|--------------------------|----------------|------------------|
| Users         | 3       | ~50 MB                   | Low            | High             |
| Vehicles      | 9       | ~200 MB                  | Medium         | Very High        |
| Inspections   | 8       | ~180 MB                  | Medium         | High             |
| Transactions  | 12      | ~250 MB                  | Medium-High    | Very High        |

**Total Estimated Index Size**: ~680 MB for 1 million documents across all collections

---

## Conclusion

These indexes are designed to optimize common query patterns while balancing write performance. Monitor actual query patterns in production and adjust indexes as needed. Use MongoDB's explain plans to validate index usage and identify missing indexes.
