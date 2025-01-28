# Date handling logic

The logic of the condition depends on how you're defining availability and how your booking system handles start dates and end dates. Let me break it down for you:



## Case 1

```
$2 < rr.end_date AND $3 > rr.start_date
```

- **$2 < rr.end_date**: Ensures the **start date** of the query is **before the end date** of an existing booking.
- **$3 > rr.start_date**: Ensures the **end date** of the query is **after the start date** of an existing booking.

This avoids overlapping bookings, but it doesn't allow bookings where **start = end**, which might work if your system doesn't allow same-day check-in and check-out.



## Case 2

```
$2 <= rr.end_date AND $3 >= rr.start_date
```

- **$2 <= rr.end_date**: Allows the query's start date to match the end date of an existing booking.
- **$3 >= rr.start_date**: Allows the query's end date to match the start date of an existing booking.

This logic makes sense **if you want to allow same-day check-out and check-in**, but it does require some clarification:

1. **Same-day check-in/check-out rule**:
   - If a user checks out on a given date, does that room become available for another user to check in on the **same day**?
   - If yes, the proposed logic is valid. If not, the current logic should remain.
2. **Edge Case: Start = End**:
   - If the user selects the same date for arrival and departure (e.g., a 1-day reservation), the room should still be considered "unavailable" for that specific day unless explicitly allowed.



## Common Sense & Best Practice:

In booking systems:

- **Check-out time** is often in the morning (e.g., 11 AM), and **check-in time** is in the afternoon (e.g., 3 PM). This ensures no overlap even if the dates are the same.
- If you're modeling availability on a full-day basis (not hour-by-hour), the proposed logic (`$2 <= rr.end_date AND $3 >= rr.start_date`) is **correct** and allows for same-day transitions.



## Example Scenarios:

#### Existing Reservation:

- Start: 2025-01-10
- End: 2025-01-12

1. Query: Start = 2025-01-12, End = 2025-01-13
   - **Case 1**: Returns `false` (not available).
   - **Case 2**: Returns `true` (available), assuming check-out is allowed same-day.
2. Query: Start = 2025-01-11, End = 2025-01-12
   - Both case 1 and case 2 would return `false` (overlapping).
3. Query: Start = 2025-01-12, End = 2025-01-12 (Same-day booking)
   - **Case 1**: Returns `true` (available).
   - **Case 2**: Returns `false` (not available), which might be more accurate depending on system rules.
