jQuery(document).ready(function() {
  jQuery("form#addUser").submit(function(event) {
    event.preventDefault();

    var error = verifyForm("addUser")

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

  jQuery("form#unsubUser").submit(function(event) {
    var error = verifyForm("unsubUser");

    if(!error) {
      return true;
    }

    event.preventDefault();
    return false;
  });
});

function verifyForm(formID) {
  jQuery("form#"+formID).siblings(".alert").remove();
  jQuery('form#'+formID+' :input.rqd').parent().removeClass('has-error');

  var error = false

  jQuery('form#'+formID+' :input.rqd').each(function() {
    var $this = jQuery(this);
    if(($this.attr('type') == 'checkbox' && !$this.is(':checked')) || jQuery(this).val() == "") {
      jQuery(this).parent().addClass('has-error');
      error = "You are missing one or more required fields.";
    }
  });

  if(!error) {
    if(/^\d{10}$/.test(jQuery('#uuidInput').val()) == false) {
      jQuery('#uuidInput').parent().addClass('has-error');
      error = "Phone numbers should only consist of 10 numbers. No spaces or symbols.";
    }
  }

  if(error) {
    jQuery("form#"+formID).parent().prepend("<div class=\"alert alert-danger\" role=\"alert\">"+error+"</alert>");
    return error;
  }

  return false
}