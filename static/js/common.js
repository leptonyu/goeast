// JavaScript Document

$(document).ready(function(){

	/*
		Home Page
			-Featured Section
			-Hover featured tool tip
	*/

	$hoverFeaturedOption = $('.hover_featured_option');			//Featured icon class name
	$hoverFeaturedOption.mousemove(function(e){
		$('body').prepend('<div class="featured_tool_tip">'+ $(this).attr('rel') +'</div>');
		$('.featured_tool_tip').fadeIn('fast').css('top',e.pageY+20).css('left',e.pageX-57);
	});

	$hoverFeaturedOption.mouseout(function(){
		$('.featured_tool_tip').fadeOut('fast',function(){
			$(this).remove();	
		});
	});

	/*
		Home Page
			-Portfolio Section
			-Hover featured portfolio item information
	*/

	$showCaption = $('.show_portfolio_caption');
	$showCaption.mouseenter(function(){
		$(this).find('.mouseover_caption').stop().animate({
			top:0
		});
	});

	$showCaption.mouseleave(function(){
		var topValue = $(this).height();
		$(this).find('.mouseover_caption').stop().animate({
			top:topValue
		});
	});
	

	/*
		Portfolio Page
			-Portfolio Section
			-Hover featured portfolio item information
	*/

	$showPFCaption = $('.portfolio_item_holder');
	$showPFCaption .mouseenter(function(){
		$(this).find('.portfolio_plus_icon').stop().animate({
			top:0
		},1);
	});

	$showPFCaption .mouseleave(function(){
		$(this).find('.portfolio_plus_icon').stop().animate({
			top:190
		},1);
	});
	
	
	/*
		Contact Page
			-Form validation
	*/
	$contact_type 	= $('#contact_type');;
	$('.submit_contact_form').click(function(){
		$name 			= $('#sender_name');
		$email 			= $('#sender_email');
		$contact_type 	= $('#contact_type');
		$website 		= $('#sender_website');
		$work_type 		= $('#project_type');
		$budget 		= $('#budget');
		$message 		= $('#message');
		
		var formValidated = true;
		var formPost = false;
		
		if(!isValidName($name.val(),'Your Name',2)){
			formValidated = false;
			$name.css('borderColor','#f66060');
		}else{
			$name.css('borderColor','');
		}
		
		if(!isValidEmail($email.val())){
			formValidated = false;
			$email.css('borderColor','#f66060');
		}else{
			$email.css('borderColor','');
		}
		
		if(!isValidURL($website.val())){
			formValidated = false;
			$website.css('borderColor','#f66060');
		}else{
			$website.css('borderColor','');
		}
		
		if($work_type.val() == 'null' && ($contact_type.val() != 'null' && $contact_type.val() == 'Hire Me')){
			formValidated = false;
			$work_type.css('borderColor','#f66060');
		}else{
			$work_type.css('borderColor','');
		}
		
		if($budget.val() == 'null' && ($contact_type.val() != 'null' && $contact_type.val() == 'Hire Me')){
			formValidated = false;
			$budget.css('borderColor','#f66060');
		}else{
			$budget.css('borderColor','');
		}
		
		if($contact_type.val() == 'null'){
			formValidated = false;
			$contact_type.css('borderColor','#f66060');
		}else{
			$contact_type.css('borderColor','');
		}
		
		if($message.val() == '' || $message.val() == 'Hey Smashy,'){
			formValidated = false;
			$message.css('borderColor','#f66060');
		}else{
			$message.css('borderColor','');
		}
		
		if(formValidated && !formPost){
			formPost = true;
			$('.submit_contact_form').attr('disabled','disabled');
			var filePath = $(this).attr('path')+'/js/send_data.php';
			$.ajax({
				type: "GET",
				url: filePath,
				data:{ name: $name.val(), email: $email.val(),contact_type: $contact_type.val(), website: $website.val(),work_type: $work_type.val(), budget: $budget.val(),message: $message.val() },
				//data: "name="+$name.val()+"&email"+$email.val()+"&contact_type"+$contact_type.val()+"&website"+$website.val()+"&work_type"+$work_type.val()+"&budget"+$budget.val()+"&message"+$message.val()
			}).done(function(data){
				$('.notification').html('<div class="alert alert-info" style="text-align:center">Inquiry successfully submitted. I will be in touch with you soon</div>');
			});	
		}else{
			$('.notification').html('<div class="alert alert-error" style="text-align:center">Some feilds needs to be filled. Checkout the feilds in red</div>');
		}
		
		
	});
	
	$contact_type.change(function(){
		if($contact_type.val() == 'Hire Me'){
			$('.project_questions').slideDown();	
			$contact_type.css('borderColor','');
		}else{
			$('.project_questions').slideUp();
			if($contact_type.val() == 'null'){
				$contact_type.css('borderColor','#f66060');
			}else{
				$contact_type.css('borderColor','');	
			}
			
		}
	});
});


function isValidName(strName,defaultValue,amount){
	var reg = /[a-zA-Z]/;

	if(reg.test(strName) == false || strName.length < amount || defaultValue == strName) 
   	{
	  	return false;
   	}
	
	return true;
}

//Validating number.
function isValidNumber(intNumber){
	var reg =/^\d{5,}$/;
	
	if (reg.test(intNumber) == false)
	{
		return false;
	}
	return true;
}

//Validating Email.
function isValidEmail(strEmail) {
   	var reg = /^([A-Za-z0-9_\-\.])+\@([A-Za-z0-9_\-\.])+\.([A-Za-z]{2,4})$/;
   	if(reg.test(strEmail) == false) 
   	{
	  	return false;
   	}
   	return true;
}


//Validating URL.
function isValidURL(strURL) {
	var reg = new RegExp(
            "^([a-zA-Z0-9\.\-]+(\:[a-zA-Z0-9\.&amp;%\$\-]+)*@)*((25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[1-9])\.(25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[1-9]|0)\.(25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[1-9]|0)\.(25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[0-9])|([a-zA-Z0-9\-]+\.)*[a-zA-Z0-9\-]+\.(com|edu|gov|int|mil|net|org|biz|arpa|info|name|pro|aero|coop|museum|[a-zA-Z]{2}))(\:[0-9]+)*(/($|[a-zA-Z0-9\.\,\?\'\\\+&amp;%\$#\=~_\-]+))*$");
   	if(reg.test(strURL) == false) 
   	{
	  	return false;
   	}
   	return true;
}



//Text feilds null on click 
function validateTextFeild(elementClassName,defaultString){
	if(elementClassName.value == ""){
		elementClassName.value = defaultString;
	}else if(elementClassName.value == defaultString){
		elementClassName.value = "";
	}
} 