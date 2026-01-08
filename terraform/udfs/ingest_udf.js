/**
 * Pub/Sub UDF for GitHub Event Ingestion
 * Parses the raw JSON payload and promotes key fields to top-level columns.
 */

function transform(inJson) {
  var inObj = JSON.parse(inJson);
  var payload = {};
  
  // Parse payload string if it exists
  if (inObj.payload) {
    try {
      if (typeof inObj.payload === 'string') {
        payload = JSON.parse(inObj.payload);
      } else {
        payload = inObj.payload;
      }
    } catch (e) {
      // Keep payload empty if parse fails
    }
  }

  var event = {
    delivery_id: inObj.delivery_id,
    signature: inObj.signature,
    received: inObj.received,
    event: inObj.event,
    payload: inObj.payload // Keep the raw payload string
  };

  // 1. Common Fields
  event.action = payload.action || null;
  event.sender = payload.sender ? payload.sender.login : null;

  if (payload.enterprise) {
    event.enterprise = payload.enterprise.name || payload.enterprise.slug || null;
  }
  
  if (payload.organization) {
    event.organization = payload.organization.login || null;
  } else if (payload.repository && payload.repository.owner) {
     // Fallback to repo owner (User accounts)
     event.organization = payload.repository.owner.login || null;
  }

  if (payload.repository) {
    event.repository = payload.repository.full_name || null;
    event.repository_language = payload.repository.language || null;
  }

  // 2. Pull Request Fields
  if (payload.pull_request) {
    event.pull_request_html_url = payload.pull_request.html_url || null;
    event.pull_request_id = payload.pull_request.id ? String(payload.pull_request.id) : null;
    event.pull_request_number = payload.pull_request.number ? String(payload.pull_request.number) : null;
    event.pull_request_merged = payload.pull_request.merged === true;
    
    if (payload.pull_request.user) {
      event.pull_request_user_login = payload.pull_request.user.login || null;
    }

    // CALCULATED FIELDS
    var adds = payload.pull_request.additions || 0;
    var dels = payload.pull_request.deletions || 0;
    var lines = adds + dels;
    event.lines_changed = String(lines);

    // Tshirt Size Calculation
    if (lines <= 0) event.tshirt_size = 'U'; // Unknown/Empty
    else if (lines < 10) event.tshirt_size = 'XS';
    else if (lines < 50) event.tshirt_size = 'S';
    else if (lines < 250) event.tshirt_size = 'M';
    else if (lines < 1000) event.tshirt_size = 'L';
    else event.tshirt_size = 'XL';

    // Duration Calculation (if closed)
    if (payload.pull_request.created_at && payload.pull_request.closed_at) {
      var created = new Date(payload.pull_request.created_at);
      var closed = new Date(payload.pull_request.closed_at);
      var duration = (closed - created) / 1000; // seconds
      if (!isNaN(duration)) {
        event.open_duration_seconds = String(Math.floor(duration));
      }
    }
  }

  // 3. Issue Fields
  // Note: PRs are also Issues in GitHub API, so this might duplicate for PRs, 
  // but useful for tracking IssueComment events specifically.
  if (payload.issue) {
    event.issue_html_url = payload.issue.html_url || null;
    event.issue_id = payload.issue.id ? String(payload.issue.id) : null;
    event.issue_number = payload.issue.number ? String(payload.issue.number) : null;
     if (payload.issue.user) {
      event.issue_user_login = payload.issue.user.login || null;
    }
  }

  // 4. Review State
  if (payload.review && payload.review.state) {
    event.review_state = payload.review.state;
  }

  return JSON.stringify(event);
}
