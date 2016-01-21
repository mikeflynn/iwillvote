jQuery(document).ready(function() {
  jQuery("form").submit(function(event) {
    event.preventDefault();
    var error = false;

    jQuery('form :input.rqd').each(function() {
      if(jQuery(this).val() == "") {
        error = true;
        jQuery(this).parent().addClass('has-error');
      }
    });

    if(error === false) {
      jQuery('form :input').each(function() {
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
          window: jQuery('#windowInput').val()
        },
        success: function(data) {
          alert("Yeah bro!");
        },
        error: function() {
          alert("There was a problem creating your account.");
        }
      });
    }

    return false;
  });
});