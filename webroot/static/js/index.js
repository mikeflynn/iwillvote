jQuery(document).ready(function() {
  jQuery("form#addUser").submit(function(event) {
    event.preventDefault();
    var error = false;

    jQuery('form#addUser :input.rqd').each(function() {
      if(jQuery(this).val() == "") {
        error = true;
        jQuery(this).parent().addClass('has-error');
      }
    });

    if(error === false) {
      jQuery('form#addUser :input').each(function() {
        jQuery(this).prop('disabled', true);
      });

      jQuery.ajax({
        url: "/api/user/add/",
        type: "POST",
        dataType: "json",
        data: {
          name: jQuery('#nameInput').val(),
          uuid: jQuery('#uuidInput').val(),
          network: jQuery('#networkInput').val(),
          state: jQuery('#stateInput').val(),
          window: jQuery('#windowInput').val(),
          landing_page: jQuery('#landingInput').val()
        },
        success: function(data) {
          jQuery("form#addUser").parent().prepend("<div class=\"alert alert-success\" role=\"alert\">You're all set! We'll remind you when it's time to vote!</alert>");
        },
        error: function() {
          jQuery("form#addUser").parent().prepend("<div class=\"alert alert-warning\" role=\"alert\">We're having trouble creating your account. Please try again later.</alert>");
        }
      });
    }

    return false;
  });
});